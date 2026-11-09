#!/usr/bin/env bash
# Usage: publish.sh <lib>
# Pushes gen/<repository> (from bin/gen-local, checked by libctl verify) to
# the lib's repo and tags the commit gen-<generator-short-sha>. CI only.
set -euo pipefail

ROOT=$(realpath "$(dirname "$0")/..")
LIB="$1"
REPO=$(yq e '.repository' - < "${ROOT}/libs/${LIB}/config.yml")
SRC="${ROOT}/${GEN_DIR:-gen}/${REPO}"
SHA="${GITHUB_SHA:?GITHUB_SHA must be set}"
TAG="gen-${SHA:0:8}"

: "${GITHUB_TOKEN:?GITHUB_TOKEN must be set to publish}"
# Token via env-scoped git config: nothing lands in remotes or config files.
export GIT_CONFIG_COUNT=1
export GIT_CONFIG_KEY_0="http.https://github.com/.extraheader"
GIT_CONFIG_VALUE_0="AUTHORIZATION: basic $(printf 'x-access-token:%s' "${GITHUB_TOKEN}" | base64 | tr -d '\n')"
export GIT_CONFIG_VALUE_0

WORK=$(mktemp -d)
trap 'rm -rf "${WORK}"' EXIT

git clone --quiet --depth 1 "https://${REPO}" "${WORK}"
git -C "${WORK}" config user.name "${GIT_COMMITTER_NAME:-jsonnet-libs-auto[bot]}"
git -C "${WORK}" config user.email "${GIT_COMMITTER_EMAIL:-jsonnet-libs-auto[bot]@users.noreply.github.com}"

find "${WORK}" -mindepth 1 -maxdepth 1 -not -name .git -exec rm -rf {} +
cp -a "${SRC}/." "${WORK}/"

# Stage first: a brand-new version dir is untracked and invisible to `git diff`.
git -C "${WORK}" add -A
if git -C "${WORK}" diff --cached --quiet; then
    echo "${LIB}: nothing to publish"
    exit 0
fi
git -C "${WORK}" commit --quiet -m "update: source github.com/jsonnet-libs-auto/generator@${SHA:0:8}"
git -C "${WORK}" tag "${TAG}"
git -C "${WORK}" push --quiet --atomic origin HEAD "refs/tags/${TAG}"
echo "${LIB}: published ${TAG}"

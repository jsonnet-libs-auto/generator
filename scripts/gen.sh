#!/usr/bin/env bash
set -euo pipefail

DIRNAME=$(realpath "$(dirname "$0")")
# Repo root holding jsonnet/ and LICENSE: the image flattens everything into
# /app, so it defaults to this script's dir; bin/gen-local sets it on the host.
GEN_ROOT=${GEN_ROOT:-${DIRNAME}}

INPUT_DIR="$1"

# Rendered fresh each run; stale skel files would otherwise leak into the output.
rm -rf "${INPUT_DIR}/skel" "${INPUT_DIR}/config.yml"
jsonnet -c -m "${INPUT_DIR}" -J "${GEN_ROOT}" -S "${INPUT_DIR}/config.jsonnet"

CONFIG_FILE="${INPUT_DIR}/config.yml"

REPO=$(yq e '.repository' - < "${CONFIG_FILE}")

OUTPUT_DIR="$2/${REPO}"

if [ "${GEN_COMMIT:-false}" != "true" ]; then
    mkdir -p "${OUTPUT_DIR}"
else
    # Token via env-scoped git config: nothing lands in remotes or config files.
    : "${GITHUB_TOKEN:?GITHUB_TOKEN must be set to publish}"
    export GIT_CONFIG_COUNT=1
    export GIT_CONFIG_KEY_0="http.https://github.com/.extraheader"
    GIT_CONFIG_VALUE_0="AUTHORIZATION: basic $(printf 'x-access-token:%s' "${GITHUB_TOKEN}" | base64 | tr -d '\n')"
    export GIT_CONFIG_VALUE_0
    git clone --depth 1 "https://${REPO}" "${OUTPUT_DIR}"

    git -C "${OUTPUT_DIR}" config user.name "${GIT_COMMITTER_NAME}"
    git -C "${OUTPUT_DIR}" config user.email "${GIT_COMMITTER_EMAIL}"
fi

# Remove everything except .git
find "${OUTPUT_DIR}" \
    -not -path "${OUTPUT_DIR}" \
    -not -path "${OUTPUT_DIR}/.git/*" \
    -not -name '.git' \
    -delete


shopt -s dotglob
cp -r "${INPUT_DIR}/skel"/* "${OUTPUT_DIR}"

k8s-gen -o "${OUTPUT_DIR}" -c "${CONFIG_FILE}"

"${DIRNAME}/docs.sh" "${INPUT_DIR}" "${OUTPUT_DIR}"


if [ "${GEN_COMMIT:-false}" != "true" ]; then
    ls -lah "${OUTPUT_DIR}"
else
    pushd "${OUTPUT_DIR}"
    # Stage first: a brand-new version dir is untracked and invisible to `git diff`.
    git add -A
    if ! git diff --cached --quiet; then
        git commit -m "update: source github.com/jsonnet-libs-auto/generator@${GITHUB_SHA:0:8}"
        # Lets consumers pin a lib to the generator commit that produced it.
        git tag "gen-${GITHUB_SHA:0:8}"
        git push --atomic origin HEAD "refs/tags/gen-${GITHUB_SHA:0:8}"
    else
        echo "Nothing to commit."
    fi
    popd
fi


if [ "${DIFF:-false}" == "true" ]; then
    echo "Diffing..."
    DIFF_DIR="${TMPDIR:-/tmp/}${REPO}"
    # HTTPS should always be usable without creds
    git clone --depth 1 "https://${REPO}" "${DIFF_DIR}" || echo "error cloning the repository!"

    (diff -r --exclude .git "${DIFF_DIR}" "${OUTPUT_DIR}" && echo "No diff!") || echo "There's a diff!"
fi

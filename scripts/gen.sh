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
    "${DIRNAME}/configure_github_ssh.sh" "${DIRNAME}"
    export GIT_SSH_COMMAND="ssh -F '${DIRNAME}/ssh/config'"
    git clone --depth 1 "ssh://git@${REPO}" "${OUTPUT_DIR}"

    git config --global user.name "${GIT_COMMITTER_NAME}"
    git config --global user.email "${GIT_COMMITTER_EMAIL}"
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
    if ! git diff --exit-code -s; then
        git add .
        git commit -m "update: source github.com/jsonnet-libs-auto/generator@${GITHUB_SHA:0:8}"
        git push
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

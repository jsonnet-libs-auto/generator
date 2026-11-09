#!/usr/bin/env bash
# Usage: gen.sh <lib-dir> <output-root>
# Renders <lib-dir>/config.jsonnet and generates the lib into
# <output-root>/<repository>. Publishing is scripts/publish.sh.
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

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

shopt -s dotglob
cp -r "${INPUT_DIR}/skel"/* "${OUTPUT_DIR}"

k8s-gen -o "${OUTPUT_DIR}" -c "${CONFIG_FILE}"

"${DIRNAME}/docs.sh" "${INPUT_DIR}" "${OUTPUT_DIR}"

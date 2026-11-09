#!/usr/bin/env bash
# Usage: changed-libs.sh [<base-ref>]
# Prints the managed libs to regenerate for changes since <base-ref>, one per
# line: only the touched libs when a change is confined to libs/<name>/, none
# for doc-only changes, every lib otherwise or when there is no base.
set -euo pipefail

ROOT=$(realpath "$(dirname "$0")/..")
cd "${ROOT}"

all=$(go run ./cmd/libctl list)
base="${1:-}"

# CHANGED_FILES (newline-separated) overrides the git diff, for testing.
if [ -n "${CHANGED_FILES:-}" ]; then
    changed="${CHANGED_FILES}"
elif [ -z "${base}" ] || ! changed=$(git diff --name-only "${base}" HEAD 2>/dev/null); then
    echo "${all}"
    exit 0
fi

docs_re='^(\.agents/|[^/]+\.md$|libs/[^/]+/AGENTS\.md$|tf/)'
shared=$(grep -vE "${docs_re}" <<< "${changed}" | grep -vE '^libs/' || true)
if [ -n "${shared}" ]; then
    echo "${all}"
    exit 0
fi

touched=$(grep -vE "${docs_re}" <<< "${changed}" | sed -nE 's#^libs/([^/]+)/.*#\1#p' | sort -u)
comm -12 <(sort <<< "${all}") <(sort <<< "${touched}")

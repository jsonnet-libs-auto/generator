---
name: update-lib
description: Bring one or more generated libs up to date with upstream releases in jsonnet-libs-auto/generator - run libctl bump, regenerate, verify and open a PR. Use when asked to update, bump or refresh a lib, or when the discover workflow's PR needs to be reproduced locally.
---

# Update a lib

Read `AGENTS.md` and `libs/<lib>/AGENTS.md` (if present) first.

1. **Discover.** `go run ./cmd/libctl discover <lib>` prints added, updated and removed outputs. An `error` means the source itself needs work: switch to the `triage-generation-failure` skill.
2. **Bump.** `go run ./cmd/libctl bump <lib>` rewrites `versions` in `libs/<lib>/upstream.json`. Never edit `versions` by hand.
3. **Generate.** `bin/gen-local <lib>`. If it fails, switch to `triage-generation-failure`.
4. **Verify.** `go run ./cmd/libctl verify <lib>`. It compares with the published repo; a `removed` list on an existing output means upstream dropped kinds within a minor: stop and report, don't work around it.
5. **Commit** `libs/<lib>/upstream.json` as `chore(libs/<lib>): bump to <newest version>`; one lib per commit.
6. **PR** with the `discover` JSON summary and the verify output's per-output `kinds` counts. A removed output (an EOL minor) must be called out in the PR body.

## Don'ts

- Don't raise `policy.minors`, add `policy.keep` or set `policy.since` to make something pass. Those are policy decisions for a human.
- Don't touch `config.jsonnet` in an update unless generation failed; that's triage.

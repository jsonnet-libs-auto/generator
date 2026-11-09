---
name: triage-generation-failure
description: Diagnose and fix a failed libctl discover/bump, bin/gen-local or libctl verify run for a lib in jsonnet-libs-auto/generator, including needs-agent issues opened by the discover workflow. Use when a lib fails to discover, generate or verify, or an upstream project changed its release layout.
---

# Triage a generation failure

Read `AGENTS.md` and `libs/<lib>/AGENTS.md` first; the quirk you need may already be written down.

## 1. Reproduce

```sh
go run ./cmd/libctl discover <lib>
bin/gen-local <lib>
go run ./cmd/libctl verify <lib>
```

Note the first failing step and its exact error.

## 2. Classify and fix

| Symptom | Likely cause | Fix |
|---|---|---|
| discover: `no upstream release matches source.tagPattern` | Tag scheme changed (prefix, `v` dropped) or repo moved | Check `git ls-remote --tags <repo>`; update `source.repo` / `source.tagPattern` |
| discover: `git ls-remote ... not found` | Repo renamed or transferred | Point `source.repo` at the new home |
| gen: `fetching <url>: 404 Not Found` | CRD file or release asset moved at some version | Find the new path at that tag; add a per-version switch in `config.jsonnet` with `jsonnet/semver.libsonnet` (see `libs/external-dns`) |
| gen: `giving up after 4 attempts ... 5xx` | Upstream outage | Retry later; not a source problem |
| gen: `no API groups found` | URL serves something without CRDs, or `prefix` no longer matches | Inspect the file; check the CRDs' `spec.group` against `prefix` |
| gen: `error unmarshaling JSON` | URL returns non-CRD YAML/HTML | Inspect the file; pick a CRD-only or bundle URL |
| verify: `files removed within an already published minor` | Upstream dropped kinds in a patch release, or the source now misses a file | If the source misses a file, fix it. If upstream really dropped them, stop and report |
| verify: `does not evaluate` | Generator or `custom/` patch bug | Fix in `pkg/` or `libs/<lib>/custom/`; add a test |

Before changing anything, search upstream issues and PRs (the CRD project and [jsonnet-libs/k8s](https://github.com/jsonnet-libs/k8s)) for the same breakage.

Prefer switching to an upstream bundle (release install manifest, `bundle.yaml`) over a longer file list.

## 3. Prove it

Rerun all three steps for the lib until `verify` passes. When the fix touches shared code (`pkg/`, `jsonnet/`, `scripts/`), also regenerate and verify every managed lib (`make libs verify`).

## 4. Record and ship

- Add or update `libs/<lib>/AGENTS.md` with the quirk: what changed upstream, from which version, and how the config handles it. One bullet, rules only.
- Commit the fix (`fix(libs/<lib>): ...`) separately from any version bump.
- In the PR, link the upstream release or commit that changed the layout. If a `needs-agent` issue triggered this, reference it with `Fixes #<n>`.

## Stop and ask a human when

- The upstream project was archived, or its CRDs moved to a different API group.
- Fixing it needs a policy change (`minors`, `keep`, `since`) or dropping a minor that is still inside the window.
- Verify reports kinds removed within a published minor and the source is correct.

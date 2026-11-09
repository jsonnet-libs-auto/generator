# Repo: jsonnet-libs-auto/generator

Generates Jsonnet libraries and publishes them to `github.com/jsonnet-libs-auto/<name>-libsonnet`. Overview: [`README.md`](README.md). Design, phases and open TODOs: [`PLAN.md`](PLAN.md); check it before starting work and tick off what you finish.

## Rules

- **No host software installs** (`mise install`, `apt`, `brew`, `pip`, `npm`, `go install`, `cargo install`, `curl | sh`, shell rc edits). Missing tool ⇒ stop and report. Tools the generator needs (jsonnet, docsonnet) are `tool` deps in `go.mod`; run them with `go tool` or `bin/gen-local`, and add new ones the same way.
- **Generate with `bin/gen-local <lib>` (or `make libs/<name>`)** into `gen/`, then `go run ./cmd/libctl verify <lib>`. A change to a lib isn't done until verify passes. Never run `scripts/publish.sh`; only CI publishes.
- **On any generator, tool, CRD or library error, search upstream issues and PRs first** (open and closed) with the GitHub MCP: the project that owns the CRDs, and [jsonnet-libs/k8s](https://github.com/jsonnet-libs/k8s), which may already have a fix to port.
- **A lib is managed iff `libs/<name>/upstream.json` exists with `"enabled": true`.** Other `libs/` dirs are upstream leftovers; they are never generated.
- **`upstream.json` `versions` is script-owned.** Change it only via `go run ./cmd/libctl bump`; edit `source` and `policy` by hand. URL layout and per-version path switches belong in `config.jsonnet`, using `v.output`, `v.tag` and `v.version`.
- **Outputs are always `major.minor`.** Never add a full-version output dir; patch releases replace the files under the same dir.
- **`policy.keep` is only for minors upstream still supports outside the window** (an LTS line). Never add one for a particular user's imports; they can pin a `gen-<sha>` tag instead.
- **Prefer one upstream bundle over hand-kept CRD file lists** (release install manifest, `bundle.yaml`, chart `crds/`): the generator only keeps `CustomResourceDefinition`s, and new CRDs then appear without edits.
- **Never hand-edit generated output.** `gen/` is local and gitignored; the `*-libsonnet` repos are written only by the publish step. Fix the source (`libs/<name>/`, `jsonnet/`, `pkg/`) and regenerate.
- **Verify the API group before adding or changing a lib.** `prefix` is the reversed `spec.group` of the CRDs; same-named upstream projects can ship different groups.
- **Never widen a version policy or delete a test to make a failing run pass.** Stop and report instead.
- **jsonnet-bundler pins need `@main`** (`jb install github.com/jsonnet-libs-auto/<name>-libsonnet@main`); jb defaults to `master`.
- **Never invoke `op`** (hangs on interactive auth). Ask the human for secrets.
- **Comments: ≤3 lines per block, ≤6 at file top**, non-obvious *why* only, no banners.
- **AGENTS.md holds rules only**; docs go in a sibling `README.md` with a one-line pointer. Per-lib quirks go in `libs/<name>/AGENTS.md`, created only when a lib has one.

## Terraform (`tf/`)

- **Allowed:** `terraform init`, `validate`, `fmt`, `plan`, `providers`, `state list`/`show`. Plan files may be written, never applied.
- **Needs explicit, single-use human approval:** `apply`, `destroy`, `import`, `state mv`/`rm`/`replace-provider`, `taint`/`untaint`, `refresh`, any `-auto-approve`.
- **One session at a time** may mutate `tf/` state.

## git-crypt

Once state is committed under git-crypt (until it moves to GCS):

- Add the `.gitattributes` pattern before `git add` of a new secret; verify with `git-crypt status -e`.
- Never remove a filter, commit `git-crypt export-key` output or `.git-crypt/keys/`, or paste decrypted content.
- Ciphertext in the working tree means the repo is locked: do not edit or commit those files.

## Git

- **Conventional Commits 1.0.0, one logical change per commit**: edit → commit → edit → commit; never edit everything and split it afterwards.
- **Message:** `<type>(<scope>)[!]: <subject>`, imperative, lowercase first letter, no trailing period, ≤72 chars. Add a body only for a *why* the subject can't carry.
- **Types:** `feat`/`fix` capability or behaviour; `chore` version bumps and regeneration; `refactor` no output change; `build` `Dockerfile`, `Makefile`, `go.mod`; `docs`, `ci`, `test`, `perf`, `revert` as specified.
- **Scopes:** `libs/<name>`, `pkg` (Go generator), `jsonnet`, `scripts`, `skills` (`.agents/skills/`), `tf`, `ci` (`.github/`), `repo` (`AGENTS.md`, `PLAN.md`, `README.md`, top-level config). Spanning several usually means two commits.
- **Allowed:** `status`, `diff`, `log`, `show`, `add`, `commit`, narrow `--amend` of an unpushed commit from this session, `merge --ff-only`, `rebase main` on your own branch (plus `--continue`/`--abort`).
- **`git push`** of the current branch to its upstream: only after the human grants it this session. Not carried across sessions, subagents or resumes.
- **Needs explicit single-use approval:** any other `rebase`; `reset --hard` or reset older than `HEAD`; `restore --worktree`; `checkout -- <path>`; `stash`; `clean`; `branch -D`; `tag -d`; `filter-branch`/`filter-repo`; `update-ref`; `symbolic-ref`; `reflog expire|delete`; any `--force`/`-f`; push of tags, `--delete`, `--mirror`, or to another remote/branch.

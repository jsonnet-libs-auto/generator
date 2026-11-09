# Plan

Hard fork of [jsonnet-libs/k8s](https://github.com/jsonnet-libs/k8s) that keeps `*-libsonnet` libraries current with as little human (and model) involvement as possible. How it works today: [`README.md`](README.md).

## Decisions

| Topic | Decision |
|---|---|
| Home | Org `jsonnet-libs-auto`; this repo is `generator`; outputs are `jsonnet-libs-auto/<name>-libsonnet`. |
| Upstream relationship | Hard fork of upstream `7fede503` (2026-10-20). Restructure freely; port upstream generator fixes by hand when useful. |
| Scope | The 14 libs the lab uses. Upstream's other `libs/` dirs stay unmanaged until they get an `upstream.json`. |
| Version policy | Per lib: newest **5** minors, each at its latest patch. `policy.keep` pins extra minors consumers import; `policy.since` sets a floor. |
| Output dirs | Always `major.minor` (jsonnet-libs mixed in full versions such as `1.27.0` and `v1.7.1`), so patch bumps keep import paths stable. |
| Version source | `git ls-remote` tags: no token, no API quota. A `helm-chart` source kind (chart tag → `appVersion`) exists but is unused. |
| CRD source | Prefer one upstream bundle per release over hand-kept file lists. |
| Merging | Fully automated on green CI; `libctl verify` is the gate. |
| Publishing | CI only, via a GitHub App token over HTTPS; each publish is tagged `gen-<generator-sha>`. |
| Repo provisioning | Terraform rendered from lib configs (`tf/main.tf.json`, committed). Local state, gitignored for now. **TODO:** git-crypt or GCS backend. |
| Docs | docsonnet markdown in each lib's `docs/`. No mkdocs or GitHub Pages. **TODO:** revisit. |
| Agent guidance | In this repo, tool-neutral: `AGENTS.md`, `libs/<name>/AGENTS.md`, `.agents/skills/`. |

## Pipeline

1. **discover / bump** (`cmd/libctl`): per managed lib, list tags, apply policy, diff against `upstream.json` `versions`, rewrite them.
2. **generate** (`bin/gen-local` → `scripts/gen.sh`): render `config.jsonnet`, run `k8s-gen` and docsonnet into `gen/`.
3. **verify** (`libctl verify`): output dirs match config, docs exist, `main.libsonnet` evaluates with every kind loaded, no kind disappears within a published minor.
4. **publish** (`scripts/publish.sh`, CI on `main`): push changed libs, tag.

Workflows: `ci` (test, then generate → verify → publish per lib), `image` (multi-arch `ghcr.io/jsonnet-libs-auto/generator`), `discover` (daily `libctl bump` → auto-merging PR; failures open a `needs-agent` issue).

The agent enters only through a `needs-agent` issue or a failing CI run, using `.agents/skills/triage-generation-failure`.

## Phases

### 1. Fork and rebrand
- [x] Import upstream source.
- [x] Point outputs at `jsonnet-libs-auto`; drop mkdocs, `requirements.txt` and gh-pages.
- [x] Host-mode generation (`bin/gen-local`) with jsonnet, docsonnet and k8s-gen pinned via `go.mod`; byte-identical to the upstream image.
- [x] Multi-arch Dockerfile with pinned bases; `image` workflow.
- [x] Only libs with `upstream.json` are generated.

### 2. `upstream.json`
- [x] `cmd/libctl` (`list`, `discover`, `bump`) with tests.
- [x] Convert the 14 lab libs. Kinds match jsonnet-libs for every overlapping minor; traefik drops `hub.traefik.io` kinds by design.
- [x] Fix the cloudnative-pg `main` bug (now uses the release manifest).
- [x] k8s builds from release tags, not branches.

### 3. Repos and first publish
- [x] Render Terraform for the managed libs; `terraform validate` passes.
- [ ] **Human:** create the GitHub App, set `PUBLISH_APP_ID` (variable) and `PUBLISH_APP_PRIVATE_KEY` (secret) on this repo, install the App on the org with Contents, Pull requests, Issues: write.
- [ ] **Human:** `terraform apply` in `tf/` (needs a token with org repo admin).
- [ ] **Human:** in repo settings, allow auto-merge and require the `ci` checks on `main`, so auto-bump PRs merge only when green.
- [ ] First publish: rerun `ci` on `main` once the App and repos exist.
- [ ] Migrate the lab: swap `tk/jsonnetfile.json` remotes, rewrite imports (`github.com/jsonnet-libs/` → `github.com/jsonnet-libs-auto/`, full-version dirs → `major.minor`), `make vendor`; review the `k8s_generated/` diff (patch-level bumps are expected now).

### 4. Scripts and tests
- [x] `libctl verify` with tests; `scripts/publish.sh` split from generation.
- [x] CI: test, then generate → verify → publish per lib.
- [x] `k8s-gen` fails on HTTP errors (with 5xx retries) and on outputs with no API groups.
- [ ] Generate only libs affected by a change (currently every lib on every push).
- [ ] testonnet golden tests for `libs/k8s/custom/` patches.

### 5. Agent guidance
- [x] Skills: `update-lib`, `add-lib`, `triage-generation-failure`.
- [x] `libs/<name>/AGENTS.md` for known quirks.
- [x] Root `AGENTS.md` rules for `upstream.json`, outputs and verify.

### 6. Runner
- [x] `discover` workflow on a daily schedule.
- [ ] Hook `needs-agent` issues and failed `ci` runs on `main` to an agent runner (kagent or similar).

### Later
- [ ] Terraform state: git-crypt or GCS.
- [ ] mkdocs / GitHub Pages, if wanted.
- [ ] Delete unmanaged upstream `libs/` dirs once nothing more is worth porting.
- [ ] external-dns 0.16–0.17 CRD path, if those minors are ever needed.
- [ ] Optionally open the same version bumps upstream.

# Plan

Hard fork of [jsonnet-libs/k8s](https://github.com/jsonnet-libs/k8s) that keeps `*-libsonnet` libraries current with as little human (and model) involvement as possible.

## Decisions

| Topic | Decision |
|---|---|
| Home | Org `jsonnet-libs-auto`; this repo is `generator`; outputs are `jsonnet-libs-auto/<name>-libsonnet`. |
| Upstream relationship | Hard fork. Restructure freely; port upstream generator fixes by hand when useful. |
| Scope | The libs the lab vendors plus `k8s`. Other upstream libs stay in `libs/` but disabled, so enabling one is a flag flip. |
| Version policy | Per lib: newest N minors (default 3), each at its latest patch. |
| Merging | Fully automated on green CI. The test suite is what makes this safe. |
| Repo provisioning | Terraform rendered from lib configs. Local state under git-crypt first; **TODO:** move to a GCS backend. |
| Docs | docsonnet markdown in each lib's `docs/`. No mkdocs or GitHub Pages; **TODO:** revisit. |
| Agent guidance | In this repo, tool-neutral: `AGENTS.md`, `libs/<name>/AGENTS.md`, `.agents/skills/`. |

## Upstream facts that shape the design

- `k8s-gen` (Go) reads a per-lib `config.yml` rendered from `libs/<name>/config.jsonnet`, fetches CRDs or an OpenAPI spec, and writes `<output>/main.libsonnet`; docsonnet writes `docs/`. `scripts/gen.sh` pushes the result to the lib repo's `main`.
- Org-specific bits are in `jsonnet/config.jsonnet` (`repository`, `site_url`, mkdocs skel), `jsonnet/github_action.jsonnet` and `jsonnet/terraform.jsonnet` (Terraform Cloud backend, `jsonnet-libs` owner).
- Version lists are hand-written inside each `config.jsonnet` in about six shapes: plain tags, `{output, version}`, `{version, chartTag}` (traefik), Kubernetes `release-X.Y` branches, explicit CRD file lists, single release assets.
- Known upstream bug: `cloudnative-pg` fetches every version's CRDs from `main`, so all its outputs are identical.
- Upstream's CI regenerates every lib on any change to shared paths and is one 138 KB generated workflow.

## Target layout

```
generator/
├── AGENTS.md
├── PLAN.md
├── README.md
├── .agents/skills/
│   ├── update-lib/SKILL.md
│   ├── add-lib/SKILL.md
│   └── triage-generation-failure/SKILL.md
├── scripts/                 # discover, bump, generate, verify, publish, configure; all --json
├── jsonnet/
│   ├── config.jsonnet       # org → jsonnet-libs-auto, mkdocs skel removed
│   └── lib.jsonnet          # turns upstream.json into specs for the common source shapes
├── libs/<name>/
│   ├── upstream.json        # what to track + script-owned version list
│   ├── config.jsonnet       # thin; imports upstream.json and jsonnet/lib.jsonnet
│   ├── AGENTS.md            # only if the lib has quirks
│   └── custom/ extensions/  # k8s only
├── tests/
├── tf/                      # generated main.tf.json
└── .github/workflows/
```

### `libs/<name>/upstream.json`

Separates what to track (edited rarely, by a human or the agent) from which versions (edited often, by scripts):

```jsonc
{
  "enabled": true,
  "source": {
    "kind": "github-release",          // github-release | github-tag | k8s-release-branch | helm-chart-tag
    "repo": "cert-manager/cert-manager",
    "tagPattern": "^v(\\d+)\\.(\\d+)\\.(\\d+)$",
    "crds": { "kind": "asset", "template": "cert-manager.crds.yaml" }
    // crds.kind: asset (release asset) | paths (explicit list, {tag} templated)
    //          | dir (list a repo dir at {tag}, filter by regex)
  },
  "policy": { "minors": 3, "outputFormat": "major.minor" },
  "prefix": "^io\\.cert-manager\\..*",
  "localName": "cert_manager",
  "versions": [{ "output": "1.21", "tag": "v1.21.0" }]
}
```

`crds.kind: dir` removes hand-maintained CRD file lists (the drift problem from [jsonnet-libs/k8s#517](https://github.com/jsonnet-libs/k8s/issues/517)).

## Pipeline

Deterministic scripts do the common path; the agent only handles exceptions.

1. **`discover`**: per enabled lib, list upstream tags/releases, apply `policy`, diff against `versions`. Emits `updates.json` (new minor, patch bump, EOL drop).
2. **`bump`**: rewrite `versions` in `upstream.json`; for `crds.kind: dir`, re-resolve the file list.
3. **`generate`**: run the pinned generator container into `gen/`. Never pushes.
4. **`verify`**:
   - every output has a non-empty `main.libsonnet` that evaluates;
   - every CRD `kind` in the fetched manifests has a generated object;
   - generated field count per version does not shrink by more than a threshold vs the published repo, unless the version was intentionally dropped;
   - `jsonnetfmt` and lint pass; docsonnet runs;
   - `k8s` only: testonnet golden tests for `custom/` patches.
5. **PR**: opened by the script with the `updates.json` summary and verify report; auto-merges on green.
6. **Publish** (on `main`): regenerate, push each changed lib to its repo's `main`, tag the lib commit `gen-<generator-short-sha>` so consumers can pin.

**Agent entry points:** a step exits non-zero, or `discover` finds a change that needs a `source` edit (renamed asset, moved CRD dir, new major with a new layout, renamed upstream repo). The runner passes the failing step's JSON to the agent, which reads `AGENTS.md`, `libs/<name>/AGENTS.md` and the `triage-generation-failure` skill, fixes the source, reruns `generate` and `verify`, records any new quirk in `libs/<name>/AGENTS.md`, and opens a PR explaining the change. The runner itself (cron, kagent, CI) is out of scope; the contract is `scripts/` plus the skills.

## Initial libs

`k8s`, `docsonnet` (doc-util, used via `github.com/jsonnet-libs/docsonnet`; decide whether to fork or keep pointing upstream), `argo-cd`, `cert-manager`, `cloudnative-pg`, `envoy-gateway`, `external-dns`, `external-secrets`, `gateway-api`, `grafana-operator`, `metallb`, `prometheus-operator`, `tailscale-operator`, `traefik`, `vertical-pod-autoscaler`, `kube-prometheus`. Known quirks to write down: traefik (chart tag → app version), grafana-operator (v4/v5 CRD path split), gateway-api (standard + experimental outputs), VPA (`vertical-pod-autoscaler-X.Y.Z` tags), argo-cd (only three minors supported upstream), k8s (`release-X.Y` branches, `custom/` and `extensions/`).

## Phases

### 1. Fork and rebrand
- [ ] Import upstream `jsonnet-libs/k8s` source (record the upstream commit in the commit body).
- [ ] Point `jsonnet/config.jsonnet` at `jsonnet-libs-auto`; drop mkdocs, `requirements.txt` and gh-pages from the skel.
- [ ] Build and publish the generator image as `ghcr.io/jsonnet-libs-auto/generator`, multi-arch, pinned base images.
- [ ] Mark every lib outside the initial list disabled.

### 2. `upstream.json`
- [ ] Write `jsonnet/lib.jsonnet` for the common source shapes.
- [ ] Convert the initial libs. Gate: for the same versions, output is byte-identical to the current `jsonnet-libs` repos.
- [ ] Fix the cloudnative-pg `main` bug.

### 3. Repos and first publish
- [ ] Render Terraform for enabled libs; local state under git-crypt. **TODO:** GCS backend.
- [ ] Create the repos (human `apply`), publish once.
- [ ] Migrate the lab: swap `tk/jsonnetfile.json` remotes, `tk-update-imports`, `make vendor`; `k8s_generated/` must stay byte-identical. Update the lab's `tk/AGENTS.md` "prefer jsonnet-libs" rule.

### 4. Scripts and tests
- [ ] `discover`, `bump`, `generate`, `verify`, `publish`, `configure`.
- [ ] `tests/` per the verify list above.
- [ ] CI: verify on PR, auto-merge on green, publish on `main`, generate only changed libs.

### 5. Agent guidance
- [ ] Skills: `update-lib`, `add-lib`, `triage-generation-failure`.
- [ ] Seed `libs/<name>/AGENTS.md` for the known quirks.
- [ ] Extend root `AGENTS.md` with the rules that now exist (versions live only in `upstream.json`, always go through `scripts/`).

### 6. Runner
- [ ] Schedule `discover`; on exceptions, spawn an agent with the matching skill.

### Later
- [ ] Terraform state → GCS.
- [ ] mkdocs / GitHub Pages, if wanted.
- [ ] Optionally open the same version bumps upstream.

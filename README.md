# generator

Generates and auto-updates `*-libsonnet` libraries from upstream Kubernetes and CRD releases. Hard fork of [jsonnet-libs/k8s](https://github.com/jsonnet-libs/k8s).

**generator** builds Jsonnet libraries for Kubernetes and operator CRDs and publishes them to `github.com/jsonnet-libs-auto/<name>-libsonnet`. It replaces upstream's hand-maintained version lists with a per-library `upstream.json`, discovers new upstream releases on a schedule, regenerates, verifies the output, and merges automatically when tests pass. An agent is only involved when an upstream layout changes or generation fails. Guidance for that agent lives in this repo (`AGENTS.md`, `libs/<name>/AGENTS.md`, `.agents/skills/`), so any runner that clones it has everything it needs.

Status and roadmap: [`PLAN.md`](PLAN.md).

## Using a lib

```bash
jb install github.com/jsonnet-libs-auto/cert-manager-libsonnet/1.21@main
```

```jsonnet
local cm = import 'github.com/jsonnet-libs-auto/cert-manager-libsonnet/1.21/main.libsonnet';
```

Each lib keeps the newest 5 minor versions of its upstream, each regenerated at the latest patch. Directories are always `major.minor`, so patch updates never change import paths. Every publish is tagged `gen-<generator-sha>` for pinning.

## Managed libs

argo-cd, cert-manager, cloudnative-pg, envoy-gateway, external-dns, external-secrets, gateway-api, grafana-operator, k8s, metallb, prometheus-operator, tailscale-operator, traefik, vertical-pod-autoscaler. The source of truth is `go run ./cmd/libctl list`.

## How a lib is defined

```
libs/<name>/
├── upstream.json    # source repo + tag pattern, version policy, tracked versions
├── config.jsonnet   # CRD/OpenAPI URLs per tracked version
└── AGENTS.md        # upstream quirks, when there are any
```

`upstream.json`:

```json
{
  "enabled": true,
  "source": { "kind": "git-tags", "repo": "https://github.com/cert-manager/cert-manager", "tagPattern": "^v(\\d+)\\.(\\d+)\\.(\\d+)$" },
  "policy": { "minors": 5 },
  "versions": [{ "output": "1.21", "tag": "v1.21.3", "version": "1.21.3" }]
}
```

- `source.kind`: `git-tags`, or `helm-chart` (chart tags, tracking the chart's `appVersion`; needs `chartPath`).
- `policy.minors`: newest minors to track. `policy.keep`: extra minors to retain. `policy.since`: oldest minor to consider.
- `versions` is written by `libctl bump`; `config.jsonnet` reads `v.output`, `v.tag` and `v.version`.

## Working on it

Needs `go` and `yq` on `PATH`; jsonnet, docsonnet and k8s-gen are built from `go.mod`.

```bash
make test                 # Go tests
make discover             # what upstream released (JSON)
make bump                 # apply it to upstream.json
make libs/cert-manager    # generate into gen/
make verify LIB=cert-manager
make tf/main.tf.json      # after adding a lib
```

CI (`.github/workflows/`): `ci` tests, generates, verifies and (on `main`) publishes every lib; `discover` opens a daily auto-merging bump PR; `image` builds `ghcr.io/jsonnet-libs-auto/generator`.

## Why a fork

Upstream is maintained on a best-effort basis ([jsonnet-libs/k8s#591](https://github.com/jsonnet-libs/k8s/issues/591)). Community version bumps routinely wait 40–140 days for review, and automating version updates has been declined upstream ([jsonnet-libs/k8s#240](https://github.com/jsonnet-libs/k8s/issues/240)). Maintainers have themselves suggested decentralising generation ([jsonnet-libs/k8s#517](https://github.com/jsonnet-libs/k8s/issues/517)).

## License

Apache 2.0, as upstream. See [`LICENSE`](LICENSE).

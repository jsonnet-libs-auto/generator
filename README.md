# generator

Generates and auto-updates `*-libsonnet` libraries from upstream Kubernetes and CRD releases. Hard fork of [jsonnet-libs/k8s](https://github.com/jsonnet-libs/k8s).

**generator** builds Jsonnet libraries for Kubernetes and operator CRDs and publishes them to `github.com/jsonnet-libs-auto/<name>-libsonnet`. It replaces upstream's hand-maintained version lists with a per-library `upstream.json`, discovers new upstream releases on a schedule, regenerates, verifies the output, and merges automatically when tests pass. An agent is only involved when an upstream layout changes or generation fails. Guidance for that agent lives in this repo (`AGENTS.md`, `libs/<name>/AGENTS.md`, `.agents/skills/`), so any runner that clones it has everything it needs.

## Status

Bootstrapping. See [`PLAN.md`](PLAN.md) for the design and phases.

## Why a fork

Upstream is maintained on a best-effort basis ([jsonnet-libs/k8s#591](https://github.com/jsonnet-libs/k8s/issues/591)). Community version bumps routinely wait 40–140 days for review, and automating version updates has been declined upstream ([jsonnet-libs/k8s#240](https://github.com/jsonnet-libs/k8s/issues/240)). Maintainers have themselves suggested decentralising generation ([jsonnet-libs/k8s#517](https://github.com/jsonnet-libs/k8s/issues/517)).

## License

Apache 2.0, as upstream. See [`LICENSE`](LICENSE).

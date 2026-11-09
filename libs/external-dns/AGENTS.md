# external-dns

- **The CRD file moved.** `docs/contributing/crd-source/crd-manifest.yaml` up to 0.15; `config/crd/standard/dnsendpoints.externaldns.k8s.io.yaml` from 0.18. 0.16 and 0.17 have neither; find their path before tracking them.
- **`keep: ["0.13"]` is a consumer pin**, not a policy choice. Drop it only when no consumer imports `0.13`.

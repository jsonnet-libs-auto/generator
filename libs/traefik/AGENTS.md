# traefik

- **Tracks `traefik/traefik` tags, not the helm chart.** CRDs come from `docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml` at the app tag, so outputs follow app minors.
- **No `hub.traefik.io` CRDs.** That file only has `traefik.io`; jsonnet-libs got hub CRDs from the chart's `crds/` dir. Add them back from `traefik/traefik-helm-chart` (kind `helm-chart`) only if someone needs Traefik Hub.

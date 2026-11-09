# envoy-gateway

- **`install.yaml` also bundles Gateway API CRDs, but the output must contain only `gateway.envoyproxy.io` kinds** (`_gen/gateway/v1alpha1`). If Gateway API kinds appear, the bundle or the generator's group handling changed; consumers should use `gateway-api-libsonnet` for those.

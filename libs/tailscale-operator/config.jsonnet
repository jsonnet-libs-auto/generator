local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='tailscale-operator',
  specs=[
    {
      output: v.output,
      crds: ['https://raw.githubusercontent.com/tailscale/tailscale/%s/cmd/k8s-operator/deploy/manifests/operator.yaml' % v.tag],
      prefix: '^com\\.tailscale\\..*',
      localName: 'tailscale-operator',
    }
    for v in upstream.versions
  ]
)

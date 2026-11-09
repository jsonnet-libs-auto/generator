local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

local release = 'https://github.com/kubernetes-sigs/gateway-api/releases/download/%s/%s-install.yaml';

config.new(
  name='gateway-api',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.k8s\\.networking\\.gateway\\..*',
      crds: [release % [v.tag, 'standard']],
      localName: 'gateway-api',
    }
    for v in upstream.versions
  ] + [
    {
      output: v.output + '-experimental',
      prefix: '^io\\.(x-)?k8s\\.networking\\.gateway\\..*',
      crds: [release % [v.tag, 'experimental']],
      localName: 'gateway-api',
    }
    for v in upstream.versions
  ]
)

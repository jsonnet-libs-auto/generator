local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='envoy-gateway',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.envoyproxy\\.gateway\\..*',
      crds: ['https://github.com/envoyproxy/gateway/releases/download/%s/install.yaml' % v.tag],
      localName: 'envoy-gateway',
    }
    for v in upstream.versions
  ]
)

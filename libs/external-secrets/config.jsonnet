local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='external-secrets',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.external-secrets\\..*',
      crds: ['https://raw.githubusercontent.com/external-secrets/external-secrets/%s/deploy/crds/bundle.yaml' % v.tag],
      localName: 'external_secrets',
    }
    for v in upstream.versions
  ]
)

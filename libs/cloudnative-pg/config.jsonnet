local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

// The release manifest carries every CRD of that release; only CRDs are used.
config.new(
  name='cloudnative-pg',
  specs=[
    {
      output: v.output,
      crds: ['https://github.com/cloudnative-pg/cloudnative-pg/releases/download/%s/cnpg-%s.yaml' % [v.tag, v.version]],
      prefix: '^io\\.cnpg\\.postgresql\\..*',
      localName: 'cloudnative-pg',
    }
    for v in upstream.versions
  ]
)

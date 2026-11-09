local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='metallb',
  specs=[
    {
      localName: 'metallb',
      output: v.output,
      prefix: '^io\\.metallb\\..*',
      crds: ['https://raw.githubusercontent.com/metallb/metallb/%s/config/manifests/metallb-native.yaml' % v.tag],
    }
    for v in upstream.versions
  ]
)

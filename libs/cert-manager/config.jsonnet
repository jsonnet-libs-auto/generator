local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='cert-manager',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.cert-manager\\..*',
      crds: ['https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.crds.yaml' % v.tag],
      localName: 'cert_manager',
    }
    for v in upstream.versions
  ]
)

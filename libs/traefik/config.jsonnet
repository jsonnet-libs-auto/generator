local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='traefik',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.traefik\\..*',
      crds: ['https://raw.githubusercontent.com/traefik/traefik/%s/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml' % v.tag],
      localName: 'traefik',
    }
    for v in upstream.versions
  ]
)

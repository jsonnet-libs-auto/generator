local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='prometheus-operator',
  specs=[
    {
      output: v.output,
      crds: ['https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/%s/bundle.yaml' % v.tag],
      prefix: '^com\\.coreos\\..*',
      localName: 'prometheus-operator',
    }
    for v in upstream.versions
  ]
)

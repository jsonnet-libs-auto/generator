local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='grafana-operator',
  specs=[
    {
      output: v.output,
      prefix: '^org\\.integreatly\\..*',
      crds: ['https://raw.githubusercontent.com/grafana/grafana-operator/%s/deploy/kustomize/base/crds.yaml' % v.tag],
      localName: 'grafana_operator',
    }
    for v in upstream.versions
  ]
)

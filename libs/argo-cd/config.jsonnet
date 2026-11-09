local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';
local manifests = ['application-crd.yaml', 'appproject-crd.yaml', 'applicationset-crd.yaml'];

config.new(
  name='argo-cd',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.argoproj\\..*',
      localName: 'argo_cd',
      crds: [
        'https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/crds/%s' % [v.tag, manifest]
        for manifest in manifests
      ],
    }
    for v in upstream.versions
  ]
)

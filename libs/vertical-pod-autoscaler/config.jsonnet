local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';

config.new(
  name='vertical-pod-autoscaler',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.k8s\\.autoscaling\\..*',
      crds: ['https://raw.githubusercontent.com/kubernetes/autoscaler/%s/vertical-pod-autoscaler/deploy/vpa-v1-crd-gen.yaml' % v.tag],
      localName: 'vertical-pod-autoscaler',
      patchDir: 'custom/core',
    }
    for v in upstream.versions
  ]
)

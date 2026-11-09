local config = import 'jsonnet/config.jsonnet';
local upstream = import 'upstream.json';
local outputs = [v.output for v in upstream.versions];

config.new(
  name='k8s',
  specs=[
    {
      output: v.output,
      openapi: 'https://raw.githubusercontent.com/kubernetes/kubernetes/%s/api/openapi-spec/swagger.json' % v.tag,
      prefix: '^io\\.k8s\\.(api|(kube-aggregator|apimachinery)\\.pkg\\.apis)\\..*',
      patchDir: 'custom/core',
      localName: 'k',
      description: 'Generated Jsonnet library for Kubernetes v' + v.output,
    }
    for v in upstream.versions
  ]
)
+ {
  'skel/README.md': (importstr './README_root.md.tmpl') % {
    version: outputs[0],
  },

  'skel/docs/README.md': (importstr './README_docs.md.tmpl') % {
    pages: std.join('\n', [
      '- [v%(version)s](%(version)s/README.md)' % { version: version }
      for version in outputs
    ]),
  },
}

local config = import 'jsonnet/config.jsonnet';
local semver = import 'jsonnet/semver.libsonnet';
local upstream = import 'upstream.json';

local crdPath(v) =
  if semver.atLeast(v.version, '0.18') then 'config/crd/standard/dnsendpoints.externaldns.k8s.io.yaml'
  else 'docs/contributing/crd-source/crd-manifest.yaml';

config.new(
  name='external-dns',
  specs=[
    {
      output: v.output,
      prefix: '^io\\.k8s\\.externaldns\\..*',
      crds: ['https://raw.githubusercontent.com/kubernetes-sigs/external-dns/%s/%s' % [v.tag, crdPath(v)]],
      localName: 'external_dns',
    }
    for v in upstream.versions
  ]
)

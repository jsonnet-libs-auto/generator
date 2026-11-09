// Run by `libctl verify` against every output; any failed assert fails verify.
local k = import 'main.libsonnet';

local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local configMap = k.core.v1.configMap;
local service = k.core.v1.service;
local servicePort = k.core.v1.servicePort;
local envVar = k.core.v1.envVar;
local role = k.rbac.v1.role;
local roleBinding = k.rbac.v1.roleBinding;
local cronJob = k.batch.v1.cronJob;

local app = container.new('app', 'img:1')
            + container.withPorts([containerPort.newNamed(8080, 'http')])
            + container.withEnvMap({ B: '2', A: '1' });
local sidecar = container.new('sidecar', 'img:2');
local cm = configMap.new('conf', { 'a.yaml': 'x: 1' });

local dep = deployment.new('foo', 2, [app, sidecar], { tier: 'web' })
            + deployment.configMapVolumeMount(cm, '/etc/conf', containers=['app']);

local podSpec = dep.spec.template.spec;

std.assertEqual(dep.apiVersion, 'apps/v1')
&& std.assertEqual(dep.kind, 'Deployment')
&& std.assertEqual(dep.spec.replicas, 2)
&& std.assertEqual(dep.spec.selector.matchLabels, { name: 'foo', tier: 'web' })
&& std.assertEqual(dep.spec.template.metadata.labels, { name: 'foo', tier: 'web' })
&& std.assertEqual(podSpec.containers[0].env, [{ name: 'A', value: '1' }, { name: 'B', value: '2' }])
&& std.assertEqual(podSpec.containers[0].ports, [{ containerPort: 8080, name: 'http' }])
&& std.assertEqual(podSpec.containers[0].volumeMounts, [{ name: 'conf', mountPath: '/etc/conf' }])
&& std.assertEqual(std.objectHas(podSpec.containers[1], 'volumeMounts'), false)
&& std.assertEqual(podSpec.volumes, [{ name: 'conf', configMap: { name: 'conf' } }])
&& std.assertEqual(std.objectHas(dep.spec.template.metadata.annotations, 'conf-hash'), true)

&& std.assertEqual(
  deployment.new('nr', null, [app]).spec + { template: null },
  { selector: { matchLabels: { name: 'nr' } }, template: null },
)

&& std.assertEqual(
  service.new('svc', { name: 'foo' }, [servicePort.newNamed('http', 80, 'http')]).spec,
  { selector: { name: 'foo' }, ports: [{ name: 'http', port: 80, targetPort: 'http' }] },
)

&& std.assertEqual(
  envVar.fromSecretRef('PASS', 'db', 'password'),
  { name: 'PASS', valueFrom: { secretKeyRef: { name: 'db', key: 'password' } } },
)

&& std.assertEqual(
  roleBinding.new('rb') + roleBinding.bindRole(role.new('r')),
  roleBinding.new('rb') + { roleRef: { name: 'r', kind: 'Role', apiGroup: 'rbac.authorization.k8s.io' } },
)

&& std.assertEqual(
  (cronJob.new('cj', '*/5 * * * *', [app]) + cronJob.emptyVolumeMount('tmp', '/tmp'))
  .spec.jobTemplate.spec.template.spec.volumes,
  [{ name: 'tmp', emptyDir: {} }],
)

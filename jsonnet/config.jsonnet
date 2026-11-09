local org = import './org.json';

{
  new(
    name,
    specs
  ):: {
    local this = self,

    suffix:: '-libsonnet',

    name:: name,
    description:: this.name + ' jsonnet library',
    repository:: 'github.com/%s/%s%s' % [org.org, this.name, this.suffix],
    branch:: org.branch,
    generator_url:: 'https://github.com/%s/%s' % [org.org, org.generator],

    'skel/LICENSE': importstr '../LICENSE',

    'config.yml': std.manifestYamlDoc({
      repository: this.repository,
      specs: [
        local proxy_port = 8001 + i;
        { proxy_port: proxy_port, openapi: 'http://localhost:%s/openapi/v2' % proxy_port, repository: this.repository }
        + specs[i]
        for i in std.range(0, std.length(specs) - 1)
      ],
    }, true),

    readme_template(name, data):: |||
      # %s Jsonnet library

      This library is generated with [`generator`](%s).

      %s
    ||| % [name, this.generator_url, data],

    'skel/README.md': this.readme_template(
      name=name,
      data=|||
        ```bash
        jb install %(repository)s/%(output)s@%(branch)s
        ```

        [Docs](docs/README.md)
      ||| % { repository: this.repository, output: specs[0].output, branch: this.branch },
    ),

    'skel/docs/README.md': this.readme_template(
      name=name,
      data=std.join('\n', [
        '- [%(output)s](%(output)s/README.md)' % spec
        for spec in specs
      ]),
    ),
  },
}

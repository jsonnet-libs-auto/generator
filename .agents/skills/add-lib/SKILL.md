---
name: add-lib
description: Onboard a new upstream project (CRDs or OpenAPI) as a managed lib in jsonnet-libs-auto/generator - choose the CRD source, write upstream.json and config.jsonnet, generate, verify and open a PR. Use when asked to add, enable or onboard a new *-libsonnet library.
---

# Add a lib

Read `AGENTS.md` first.

## 1. Reuse what exists

If `libs/<name>/config.jsonnet` already exists (inherited from jsonnet-libs), start from it: it records the CRD URLs and `prefix` that worked upstream. Rewrite it to read `upstream.json` like the managed libs (`libs/cert-manager/config.jsonnet` is the simplest example).

## 2. Find the source

- **Versions:** the git repo whose semver tags mark releases. Confirm with `git ls-remote --tags --refs <repo>` and write a `tagPattern` capturing major, minor and patch. Use `kind: helm-chart` (with `chartPath`) only if the chart repo is the only place CRDs live.
- **CRDs, in order of preference:**
  1. a release asset or repo file bundling every CRD (install manifest, `bundle.yaml`, `crds.yaml`): non-CRD documents are ignored;
  2. a stable CRD directory, listed in `config.jsonnet`;
  3. per-file URLs, as a last resort.
- Check the URL exists at the newest and the oldest tag the policy will select.

## 3. Write the lib

`libs/<name>/upstream.json`:

```json
{
  "enabled": true,
  "source": { "kind": "git-tags", "repo": "https://github.com/<org>/<repo>", "tagPattern": "^v(\\d+)\\.(\\d+)\\.(\\d+)$" },
  "policy": { "minors": 5 },
  "versions": []
}
```

- `policy.minors` is 5 unless the human said otherwise.
- `policy.since` only when old releases have no CRDs; `policy.keep` only for a minor a consumer is known to import.

`libs/<name>/config.jsonnet`: a spec per `upstream.versions` entry with `output: v.output`, URLs from `v.tag` / `v.version`, and:

- `prefix`: the reversed `spec.group` of the CRDs (`cert-manager.io` → `^io\\.cert-manager\\..*`). Read `spec.group` from the actual CRD, don't guess.
- `localName`: the lib name in snake case, used in docs examples.

## 4. Fill, generate, verify

```sh
go run ./cmd/libctl bump <name>
bin/gen-local <name>
go run ./cmd/libctl verify <name>
make tf/main.tf.json
```

Check `gen/github.com/jsonnet-libs-auto/<name>-libsonnet/<newest>/_gen/` lists the kinds you expect.

## 5. Ship

- One commit for the lib (`feat(libs/<name>): add <name>`), one for `tf/main.tf.json` (`feat(tf): add <name>-libsonnet repo`).
- PR body: source repo, CRD URL pattern, selected versions, and the kinds generated for the newest version.
- The repo only exists after a human runs `terraform apply`; publishing fails until then. Say so in the PR.
- Write `libs/<name>/AGENTS.md` only if you hit a quirk.

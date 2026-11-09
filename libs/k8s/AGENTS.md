# k8s

- **Generated from `api/openapi-spec/swagger.json` at the release tag**, not the `release-X.Y` branch jsonnet-libs used, so output is reproducible and moves with patch releases.
- **`custom/core/` is merged into every version.** A patch that breaks on a new Kubernetes minor must be fixed there, not by dropping the version.
- **README templates** (`README_root.md.tmpl`, `README_docs.md.tmpl`) replace the generic skel README.
- **`tests/custom_patches.jsonnet` pins the `custom/` constructors' output** and runs against every version in `libctl verify`. Change a patch only with a matching test change.

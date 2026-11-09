# cloudnative-pg

- **CRDs come from the release manifest `cnpg-<version>.yaml`**, so new CRDs are picked up without a file list. jsonnet-libs read every version from `main`, so all its outputs were the same schema.

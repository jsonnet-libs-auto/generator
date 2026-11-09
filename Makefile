IMAGE ?= ghcr.io/jsonnet-libs-auto/generator
IMAGE_TAG ?= dev

LIBCTL := go run ./cmd/libctl
# Managed libs: those with libs/<name>/upstream.json and "enabled": true.
LIBS = $(shell $(LIBCTL) list)
IMPORTS = $(foreach lib,$(LIBS),(import 'libs/$(lib)/config.jsonnet'),)

OBJECTMETA_VERSION := v1.23.4

.DEFAULT_GOAL := help

help:
	@echo "make test            Go tests"
	@echo "make verify LIB=<name> check a generated lib before publishing"
	@echo "make libs            generate every managed lib into gen/"
	@echo "make libs/<name>     generate one lib into gen/"
	@echo "make discover        report upstream version changes (JSON)"
	@echo "make bump            apply upstream version changes to upstream.json"
	@echo "make tf/main.tf.json render Terraform for the lib repos"
	@echo "make image           build the generator image"

test:
	go vet ./...
	go test ./...

libs:
	bin/gen-local $(LIBS)

libs/%:
	bin/gen-local $*

discover:
	@$(LIBCTL) discover

bump:
	@$(LIBCTL) bump

.PHONY: tf/main.tf.json
tf/main.tf.json:
	go tool jsonnet -c -m . -S -J . --tla-code "libs=[$(IMPORTS)]" jsonnet/terraform.jsonnet

update_objectmeta:
	curl -sL "https://raw.githubusercontent.com/kubernetes/kubernetes/$(OBJECTMETA_VERSION)/api/openapi-spec/swagger.json" | go tool jsonnet -o pkg/swagger/objectmeta.json jsonnet/objectmeta.jsonnet

image:
	docker build -t $(IMAGE):$(IMAGE_TAG) .

.PHONY: help test libs discover bump verify update_objectmeta image

verify:
	@$(LIBCTL) verify $(or $(LIB),$(LIBS))

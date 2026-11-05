default: testacc

.PHONY: testacc run setup fmt fmt-tf-conf lint lint-fix \
        sdk sdk-os bump-oas bump-oas-os \
        generate-spec generate-os-spec generate-framework \
        generate-resource-registry check-registry-drift \
        scaffold-init scaffold \
        check-scaffold-drift fix-scaffold-drift audit codegen-canary \
        new-feature docs

API_URL           ?= https://api.eu-west-2.numspot.com
OS_URL            ?= https://objectstorage.eu-west-2.numspot.com
API_OAS_PATH      ?= ./internal/sdk/api/public-oas.yaml
OS_OAS_PATH       ?= ./internal/sdk/objectstorage/public-object-storage.yaml
API_CONF          ?= ./internal/sdk/api/sdk.conf.yaml
OS_CONF           ?= ./internal/sdk/objectstorage/sdk-os.conf.yaml
# HashiCorp codegen CLIs are tech-preview and effectively unmaintained
# (last releases: openapi v0.3.0, framework v0.4.1 in 2024). They are pinned to
# exact versions so generation stays reproducible and a future upstream change
# cannot silently break our build. Bump deliberately, not via @latest.
CODEGEN_OPENAPI   := github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@v0.3.0
CODEGEN_FRAMEWORK := github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@v0.4.1
OAPI_CODEGEN      := github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1

# Local tooling. terrafmt and tfplugindocs are managed as go tool dependencies
# in go.mod. golangci-lint is installed into bin/ via `go install` so it is
# built with the local Go toolchain (avoids prebuilt-binary Go version skew).
BIN_DIR                := $(shell pwd)/bin
GOLANGCI_LINT_VERSION  := v2.13.2
GOLANGCI_LINT          := $(BIN_DIR)/golangci-lint
TERRAFMT               := go tool terrafmt
TFPLUGINDOCS           := go tool tfplugindocs
export PATH            := $(BIN_DIR):$(PATH)

# ── Build & Test ────────────────────────────────────────────────────

testacc:
	TF_ACC=1 VCR_MODE=replay go test -tags acc ./internal/test -v $(TESTARGS) -timeout 120m

record-cassettes:
	TF_ACC=1 VCR_MODE=record go test -tags acc ./internal/test -v $(TESTARGS) -timeout 120m

run:
	go run main.go

# ── Formatting & Lint ───────────────────────────────────────────────

setup: $(GOLANGCI_LINT)
	@echo "golangci-lint $(GOLANGCI_LINT_VERSION) installed to $(BIN_DIR)"
	@echo "terrafmt and tfplugindocs are fetched by 'go tool' on first use"

fmt: fmt-tf-conf $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) fmt

fmt-tf-conf:
	find . | egrep "test.go" | sort | while read f; do $(TERRAFMT) fmt -fv $$f; done
	terraform fmt -recursive examples/

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix --path-mode abs

$(GOLANGCI_LINT):
	@mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# ── SDK Generation ──────────────────────────────────────────────────
# Regenerate the Go SDK client from the pinned OAS snapshot in the repo.
# OAS snapshots are updated deliberately via: make bump-oas

sdk:
	go run $(OAPI_CODEGEN) --config=$(API_CONF) $(API_OAS_PATH)

sdk-os:
	go run $(OAPI_CODEGEN) --config=$(OS_CONF) $(OS_OAS_PATH)

# ── OAS Snapshot Update ───────────────────────────────────────────
# Download the latest OAS from the live API and replace the pinned snapshot.
# Review the diff before committing. Run `make sdk` afterwards.

bump-oas:
	@echo "Downloading API OAS from $(API_URL)/openapi ..."
	@mkdir -p $(dir $(API_OAS_PATH))
	curl -sSf --compressed $(API_URL)/openapi > $(API_OAS_PATH)
	@OAS_VERSION=$$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['info']['version'])" $(API_OAS_PATH)) && \
	TODAY=$$(date +%Y-%m-%d) && \
	sed -i '' "s#^| API *|.*#| API            | $$OAS_VERSION | $$TODAY |#" VERSION.md && \
	echo "OAS snapshot updated (API version: $$OAS_VERSION). Run 'make sdk' to regenerate the SDK client." && \
	echo "Review the diff before committing: git diff $(API_OAS_PATH) VERSION.md"

bump-oas-os:
	@echo "Downloading Object Storage OAS from $(OS_URL)/openapi ..."
	@mkdir -p $(dir $(OS_OAS_PATH))
	curl -sSf --compressed $(OS_URL)/openapi > $(OS_OAS_PATH)
	@OAS_VERSION=$$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['info']['version'])" $(OS_OAS_PATH)) && \
	TODAY=$$(date +%Y-%m-%d) && \
	sed -i '' "s#^| Object Storage *|.*#| Object Storage | $$OAS_VERSION | $$TODAY |#" VERSION.md && \
	echo "OAS snapshot updated (Object Storage version: $$OAS_VERSION). Run 'make sdk-os' to regenerate the SDK client." && \
	echo "Review the diff before committing: git diff $(OS_OAS_PATH) VERSION.md"

# ── TF Schema Generation ───────────────────────────────────────────
# Generate the Terraform schema JSON from the OAS spec, then generate
# the framework resource/datasource Go scaffolding.
#
# Usage: make generate-spec RESOURCE_NAME=vpc
#        make generate-framework RESOURCE_NAME=vpc

SPEC_YAML  := internal/services/$(RESOURCE_NAME)/$(RESOURCE_NAME)_spec.yaml
SPEC_JSON  := internal/services/$(RESOURCE_NAME)/$(RESOURCE_NAME)_spec.json
DESCRIPTOR := internal/services/$(RESOURCE_NAME)/descriptor.yaml

generate-spec:
	go run ./cmd/inject-ignores/ $(SPEC_YAML) > $(SPEC_YAML).injected && \
	go run $(CODEGEN_OPENAPI) generate \
		--config=$(SPEC_YAML).injected \
		--output=$(SPEC_JSON) \
		$(API_OAS_PATH); \
	rm -f $(SPEC_YAML).injected
	go run ./cmd/inject-schema-overrides/ $(SPEC_JSON) $(DESCRIPTOR)

generate-os-spec:
	go run ./cmd/inject-ignores/ $(SPEC_YAML) > $(SPEC_YAML).injected && \
	go run $(CODEGEN_OPENAPI) generate \
		--config=$(SPEC_YAML).injected \
		--output=$(SPEC_JSON) \
		$(OS_OAS_PATH); \
	rm -f $(SPEC_YAML).injected
	go run ./cmd/inject-schema-overrides/ $(SPEC_JSON) $(DESCRIPTOR)

generate-framework: generate-resource-framework generate-datasource-framework

generate-resource-framework:
	go run $(CODEGEN_FRAMEWORK) generate \
		resources --input internal/services/$(RESOURCE_NAME)/$(RESOURCE_NAME)_spec.json \
		--output internal/services/$(RESOURCE_NAME)/

generate-datasource-framework:
	go run $(CODEGEN_FRAMEWORK) generate \
		data-sources --input internal/services/$(RESOURCE_NAME)/$(RESOURCE_NAME)_spec.json \
		--output internal/services/$(RESOURCE_NAME)/

# ── Scaffold ────────────────────────────────────────────────────────
# Create a new resource/data-source using the code generator.
#
# Usage: make scaffold-init RESOURCE_NAME=organisation_role
#          (edit descriptor.yaml, set go_name, singular, plural, flags)
#        make scaffold RESOURCE_NAME=organisation_role
#          (generates core + resource + datasource + patches provider.go)

scaffold-init:
	@mkdir -p internal/services/$(RESOURCE_NAME)
	@printf '# yaml-language-server: $$schema=../../../docs/internal/descriptor.schema.json\nname: %s\ngo_name:\nsingular:\nplural:\nscope: space\nhas_tags: false\nis_async: false\nhas_update: false\ncreate_direct: true\nhas_resource: true\nhas_datasource: true\nscaffold_check: false\nreason: ""\n' "$(RESOURCE_NAME)" > internal/services/$(RESOURCE_NAME)/descriptor.yaml
	@echo "Created internal/services/$(RESOURCE_NAME)/descriptor.yaml"
	@echo "Set scaffold_check: true once generated files are stable, or keep false and fill in reason."

scaffold:
	go run ./cmd/scaffold gen --descriptor internal/services/$(RESOURCE_NAME)/descriptor.yaml

check-scaffold-drift:
	go run ./cmd/scaffold drift

fix-scaffold-drift:
	go run ./cmd/scaffold drift -fix

audit:
	go run ./cmd/scaffold audit

# ── Resource Registry ───────────────────────────────────────────────
# Regenerate, from docs/internal/resources-registry.yaml:
#   - docs/internal/RESOURCE_REGISTRY.md (async/retry reference table)
#   - OWNERS.md                          (human-readable team -> resources map)
#   - .github/CODEOWNERS                 (functional GitHub path -> team owners)
generate-resource-registry:
	go run ./cmd/resource-registry

# Fail if any registry-derived file is stale (not regenerated + committed).
# Mirrors the regenerate-and-diff guard used by the go_generate CI job.
check-registry-drift:
	go run ./cmd/resource-registry
	git diff --exit-code -- docs/internal/RESOURCE_REGISTRY.md OWNERS.md .github/CODEOWNERS \
		|| (echo "Registry-derived files are stale. Run 'make generate-resource-registry' and commit."; exit 1)

new-feature:
	mkdir -p internal/services/$(RESOURCE_NAME)
	touch internal/services/$(RESOURCE_NAME)/$(RESOURCE_NAME)_spec.yaml
	echo "package core" > internal/core/$(RESOURCE_NAME).go
	go run $(CODEGEN_FRAMEWORK) scaffold data-source \
		--name $(RESOURCE_NAME) \
		--output-dir internal/services/$(RESOURCE_NAME)/ \
		--package $(RESOURCE_NAME)
	go run $(CODEGEN_FRAMEWORK) scaffold resource \
		--name $(RESOURCE_NAME) \
		--output-dir internal/services/$(RESOURCE_NAME)/ \
		--package $(RESOURCE_NAME)

docs:
	$(TFPLUGINDOCS) generate --provider-name numspot
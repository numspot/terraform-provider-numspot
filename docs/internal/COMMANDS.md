# Developer commands

Reference for the developer tooling in this repo: the `make` targets (see the
[GNUmakefile](../../GNUmakefile)) and the `cmd/scaffold` subcommands (see
[cmd/scaffold/main.go](../../cmd/scaffold/main.go)). For the feature workflow that
chains these together, see [ADDING_A_FEATURE.md](ADDING_A_FEATURE.md); for the
descriptor fields, see [DESCRIPTOR.md](DESCRIPTOR.md).

Most generation targets take `RESOURCE_NAME=<folder>` (the service folder under
`internal/services/`, e.g. `vpc`, `dhcpoptions`).

## Build, test, format, lint

| Command                           | What it does                                                                        |
| --------------------------------- | ----------------------------------------------------------------------------------- |
| `make testacc`                    | Acceptance tests in VCR **replay** mode (`TF_ACC=1 VCR_MODE=replay`, 120m timeout). |
| `make record-cassettes`           | Acceptance tests in VCR **record** mode (hits the live API).                        |
| `make run`                        | `go run main.go` (provider entrypoint).                                             |
| `make setup`                      | Install `golangci-lint` into `bin/`; `terrafmt`/`tfplugindocs` stay `go tool` deps. |
| `make fmt`                        | `gofumpt` + `gci` on Go, `terrafmt`/`terraform fmt` on examples and test HCL.       |
| `make lint`                       | `golangci-lint` (pinned, installed via official script into `bin/`).                |
| `make lint-fix`                   | `golangci-lint --fix`.                                                              |
| `go build ./...` / `go vet ./...` | Standard build / vet.                                                               |

## SDK generation

| Command            | What it does                                                                                                                    |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| `make sdk`         | Regenerate the Go SDK client from the pinned API OAS snapshot.                                                                  |
| `make sdk-os`      | Regenerate the Object Storage SDK client from its pinned OAS snapshot.                                                          |
| `make bump-oas`    | Download the latest API OAS from the live API, replace the snapshot, update `VERSION.md`. Review the diff, then run `make sdk`. |
| `make bump-oas-os` | Same, for the Object Storage OAS (then `make sdk-os`).                                                                          |

The codegen CLIs (`tfplugingen-openapi`, `tfplugingen-framework`, `oapi-codegen`)
are tech-preview and **pinned** to exact versions in the GNUmakefile. Never bump
them to `@latest`.

## Schema + scaffold generation

| Command                                   | What it does                                                                                                          |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `make scaffold-init RESOURCE_NAME=x`      | Create `internal/services/x/` with a starter `descriptor.yaml` (carrying the JSON-schema modeline) and `x_spec.yaml`. |
| `make generate-spec RESOURCE_NAME=x`      | Generate `x_spec.json` from the OAS + spec YAML (inject-ignores → tfplugingen-openapi → inject-schema-overrides).     |
| `make generate-os-spec RESOURCE_NAME=x`   | Same as `generate-spec` but against the Object Storage OAS.                                                           |
| `make generate-framework RESOURCE_NAME=x` | Generate the `*_gen.go` files (schema + models) from `x_spec.json`.                                                   |
| `make scaffold RESOURCE_NAME=x`           | Generate core + resource + datasource and patch `provider.go` (`scaffold gen` under the hood).                        |

## Audit + drift

These commands are run together in CI, but they check different things.

| Command                     | What it does                                                                                           |
| --------------------------- | ------------------------------------------------------------------------------------------------------ |
| `make audit`                | Check every `descriptor.yaml` for well-formedness and completeness (no regeneration).                  |
| `make check-scaffold-drift` | Regenerate guarded resources and fail if any committed file differs.                                   |
| `make fix-scaffold-drift`   | Detect drifted `scaffold_check`/`spec_check` resources and regenerate only the artifacts that drifted. |

### `make audit` — is the descriptor clean?

`go run ./cmd/scaffold audit` reads every `descriptor.yaml` and checks that it is
**well-formed and complete**. It does **not** regenerate files and it does **not**
compare against committed output.

It fails when:

- a `scaffold_check: true` descriptor still has a residual `TODO` mapping
- a `mapping_overrides` entry is missing a `reason`
- the service folder name does not match the concatenated-lowercase `name`
- the spec files are not named `<folder>_spec.{json,yaml}`
- core function names inferred from the descriptor do not match the actual code

Use it locally after editing a descriptor, **before** committing generated files.

### `make check-scaffold-drift` — are generated files up to date?

`go run ./cmd/scaffold drift` **regenerates** files and compares them
byte-for-byte with what is committed.

| Descriptor flag        | What is regenerated and compared                                                                                     |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `scaffold_check: true` | `resource_*.go`, `datasource_*.go`, `internal/core/*.go`                                                             |
| `spec_check: true`     | `<name>_spec.json` (full pipeline: OAS → spec YAML → inject-ignores → tfplugingen-openapi → inject-schema-overrides) |

It fails as soon as any generated file differs from the committed version.

### `make fix-scaffold-drift` — fix drift in one shot

`go run ./cmd/scaffold drift -fix` detects which guarded descriptors have
scaffold or spec drift and regenerates only the artifacts that actually drifted:

| Failed check          | Targets invoked                                                                                                                 |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `scaffold_check` only | `make scaffold RESOURCE_NAME=<name>`                                                                                            |
| `spec_check` only     | `make generate-spec RESOURCE_NAME=<name>` (or `generate-os-spec` for `bucket`) + `make generate-framework RESOURCE_NAME=<name>` |
| both                  | `generate-spec` (or `generate-os-spec`) → `generate-framework` → `scaffold`                                                     |

It stops on the first failing resource/target so the error is easy to find.

### Typical verification flow

```bash
# 1. Make sure the descriptor itself is valid
make audit

# 2. Regenerate everything for the resource you changed
make generate-spec RESOURCE_NAME=x
make generate-framework RESOURCE_NAME=x
make scaffold RESOURCE_NAME=x

# 3. Make sure there is no drift left
make check-scaffold-drift

# If multiple resources drifted, fix them automatically and re-check
make fix-scaffold-drift
make check-scaffold-drift
```

### CI

The GitHub Actions CI workflow runs these checks on every pull request and push to `main`:

```bash
make audit                  # descriptor quality
make check-scaffold-drift   # generated-file drift guard
make check-registry-drift   # RESOURCE_REGISTRY.md / OWNERS.md / .github/CODEOWNERS
```

## Resource registry

| Command                           | What it does                                                                                                                                                                     |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `make generate-resource-registry` | Regenerate [RESOURCE_REGISTRY.md](RESOURCE_REGISTRY.md) from [resources-registry.yaml](resources-registry.yaml) (`go run ./cmd/resource-registry`). Edit the YAML, never the MD. |

## `cmd/scaffold` subcommands (raw)

The `make` targets above wrap these. Run directly for one-off flags:

```bash
go run ./cmd/scaffold gen -descriptor internal/services/<name>/descriptor.yaml [-dry-run]
go run ./cmd/scaffold audit [-descriptor <path>]   # all descriptors when -descriptor is omitted
go run ./cmd/scaffold drift                          # all scaffold_check / spec_check descriptors
```

- `gen` generates the core/resource/datasource files and patches `provider.go`.
  `-dry-run` prints the generated code without writing files (used by `drift`).
- `audit` reports naming/override/convention issues and fails when a
  `scaffold_check: true` descriptor still produces a residual `TODO`.
- `drift` regenerates and diffs: the `.go` files (when `scaffold_check: true`) and
  the `*_spec.json` (when `spec_check: true`), failing on any difference.

## Legacy

| Command                            | What it does                                                                                |
| ---------------------------------- | ------------------------------------------------------------------------------------------- |
| `make new-feature RESOURCE_NAME=x` | HashiCorp codegen scaffolding (no `BaseResource`). **Prefer `scaffold-init` + `scaffold`.** |

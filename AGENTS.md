# AGENTS.md

Guidance for AI agents (Cursor, OpenCode, and any other tool that reads a
root `AGENTS.md`) working in `terraform-provider-numspot`. Humans are welcome
too. This file is intentionally short and link-first: the authoritative detail
lives in [`docs/internal/`](docs/internal/).

## Project overview

This is the Terraform provider for [NumSpot Cloud](https://numspot.com/). Most
of the work is adding or maintaining resources and data sources, which are
**generated** from a per-resource descriptor and then refined by hand.

- Each resource/data source lives in `internal/services/<name>/` (the folder is
  the resource name with underscores removed, e.g. `dhcp_options` ->
  `dhcpoptions`).
- Generation pipeline: `descriptor.yaml` + `<name>_spec.yaml` -> `<name>_spec.json`
  (IR) -> `*_gen.go` (schema + models) -> scaffolded `core` + `resource` +
  `datasource` business logic.
- Runtime call flow:

```
resource_<name>.go    (CRUD orchestration, serialize/deserialize)
    -> internal/core/<name>.go    (API calls)
        -> internal/sdk/api       (oapi-codegen client, generated from OAS)
            -> NumSpot API
```

Full workflow: [docs/internal/ADDING_A_FEATURE.md](docs/internal/ADDING_A_FEATURE.md).
Descriptor fields: [docs/internal/DESCRIPTOR.md](docs/internal/DESCRIPTOR.md).

## Key commands

Most generation targets take `RESOURCE_NAME=<folder>` (e.g. `vpc`, `dhcpoptions`).
Full list: [docs/internal/COMMANDS.md](docs/internal/COMMANDS.md).

| Command                                   | What it does                                                                    |
| ----------------------------------------- | ------------------------------------------------------------------------------- |
| `make fmt`                                | `gofumpt` + `gci` on Go, `terrafmt` on HCL.                                     |
| `go build ./...` / `go vet ./...`         | Build / vet.                                                                    |
| `make audit`                              | `cmd/scaffold audit` — naming, override `reason`s, residual `TODO`s.            |
| `make check-scaffold-drift`               | Verify guarded descriptors match generated output.                              |
| `make scaffold-init RESOURCE_NAME=x`      | Create the service folder + starter `descriptor.yaml` + `x_spec.yaml`.          |
| `make generate-spec RESOURCE_NAME=x`      | Generate `x_spec.json` (IR) from OAS + spec YAML.                               |
| `make generate-framework RESOURCE_NAME=x` | Generate the `*_gen.go` schema + models.                                        |
| `make scaffold RESOURCE_NAME=x`           | Generate core + resource + datasource and patch `provider.go`.                  |
| `make sdk`                                | Regenerate the Go SDK client from the pinned OAS snapshot.                      |
| `make generate-resource-registry`         | Regenerate `RESOURCE_REGISTRY.md` / `OWNERS.md` from `resources-registry.yaml`. |

The codegen CLIs (`tfplugingen-openapi`, `tfplugingen-framework`, `oapi-codegen`)
are tech-preview and **pinned** in the [GNUmakefile](GNUmakefile). Never bump
them to `@latest`.

## Do-not-edit files

These are generated; edit the **source** and rerun the matching target instead.

| File                                                                                              | Source of truth                                                                | Regenerate with                   |
| ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ | --------------------------------- |
| `internal/services/*/**/*_gen.go`                                                                 | `*_spec.json`                                                                  | `make generate-framework`         |
| `internal/services/*/*_spec.json` (IR)                                                            | OAS + `*_spec.yaml` + `descriptor.yaml`                                        | `make generate-spec`              |
| `internal/sdk/api/*`                                                                              | OAS snapshot                                                                   | `make sdk`                        |
| [docs/internal/RESOURCE_REGISTRY.md](docs/internal/RESOURCE_REGISTRY.md) / [OWNERS.md](OWNERS.md) | [docs/internal/resources-registry.yaml](docs/internal/resources-registry.yaml) | `make generate-resource-registry` |

The scaffolded `core/<name>.go`, `resource_<name>.go`, and `datasource_<name>s.go`
**are** editable (review residual `TODO`s after `make scaffold`), but resources
with `scaffold_check: true` must stay in sync with the descriptor (drift-guarded).

> **Linting generated files.** Do not hand-edit generated files to satisfy a
> linter. Add a targeted exclusion to `.golangci.yml` (e.g. a `path` rule under
> `linters.exclusions.rules`) and fix the generator/template or the descriptor
> instead. When a linter flag is caused by generated code that scaffolded files
> simply call (e.g. terraform-plugin-framework constructors using
> `context.Background()`), the same targeted exclusion may be applied to the
> affected scaffolded files and `helpers.go` rather than sprinkling `nolint`
> comments across generated call sites.

## Skills / common tasks

Tool-agnostic recipes. Each links to the authoritative doc.

- **Add a resource or data source.** Follow the 8-step flow in
  [ADDING_A_FEATURE.md](docs/internal/ADDING_A_FEATURE.md):
  ```bash
  make scaffold-init RESOURCE_NAME=x
  # edit descriptor.yaml + x_spec.yaml
  make generate-spec RESOURCE_NAME=x
  make generate-framework RESOURCE_NAME=x
  make scaffold RESOURCE_NAME=x
  # fill residual TODOs, then verify
  make fmt && go build ./... && make audit
  ```
- **Tune a field mapping.** Prefer descriptor pattern flags (`id_source`,
  `preserve_from_state`, `exclude_from_request`, `computed_fields`,
  `null_on_read`, ...) over hand edits; drop to `mapping_overrides` (each needs a
  `reason`) only when a flag can't express it. See
  [DESCRIPTOR.md](docs/internal/DESCRIPTOR.md) and the `helpers.go` section below.
- **Regenerate the SDK / bump OAS.** `make sdk` (or `make bump-oas` then review
  the diff and `make sdk`). Object Storage: `make sdk-os` / `make bump-oas-os`.
- **Register a resource / update owners.** Edit
  [docs/internal/resources-registry.yaml](docs/internal/resources-registry.yaml),
  then `make generate-resource-registry`. Never hand-edit `RESOURCE_REGISTRY.md`
  or `OWNERS.md`.
- **Verify before committing.** `make fmt && go build ./... && make audit`, plus
  `make check-scaffold-drift` when touching a `scaffold_check`/`spec_check`
  resource (including lint fixes in scaffolded `resource_*.go` or
  `datasource_*.go` files). If drift is reported, run `make fix-scaffold-drift`
  and re-check before committing.

## `helpers.go` convention

When the scaffold field mapper can't auto-map a field, the per-resource logic
lives in a hand-written `internal/services/<name>/helpers.go`. This keeps the
generated files clean while still letting the descriptor drive generation.

- **What it is.** An optional, hand-written file in the **same package** as the
  generated `resource_<name>.go` / `datasource_<name>s.go`. It is **not**
  generated and **not** overwritten by `make scaffold`, so it is the right home
  for logic the generator can't express.
- **When to add one.** Add a helper when a field needs real Go logic, then
  reference it from `descriptor.yaml` via:
  - `mapping_overrides.serialize` / `deserialize` / `datasource` /
    `datasource_params` using an `expression`, e.g.
    `expression: types.StringValue(getVpcId(http.VpcToVirtualGatewayLinks))`;
  - `core.create.extra_args` / `core.delete.extra_args` (request/link body
    builders threaded into the core call);
  - `core.update.calls[].args` (update-body builders and `plan` vs `state` diff
    helpers).
- **What goes in it.** Small, focused functions:
  - derived serialize getters — `firstPublicIpId`, `linkVmFromVolume`,
    `getVpcId`, `securityGroupIdsFromVm`;
  - set/list converters — `backendIps`, `azNamesToTfList`, `stringSetToPtrList`;
  - create/update body builders — `deserializeUpdateVm`,
    `deserializeUpdateLoadBalancer`, `createListenersBody`.
    Prefer the shared converters in `internal/utils` inside helpers
    (`utils.FromTfStringSetToStringList`, `utils.FromTfInt64ToIntPtr`,
    `utils.TimeToTfString`, ...) rather than re-implementing conversions.
- **Signature conventions** (from the real examples below). Helpers typically
  take `ctx context.Context`, the TF model (e.g. `resource_vm.VmModel`) or an SDK
  type (e.g. `*api.Vm`), and `diags *diag.Diagnostics`; they return TF types
  (`types.String`, `types.Set`, ...) or `api.*` request bodies. Be nil-safe on
  pointers and return typed nulls (`types.StringNull()`) rather than panicking.
- **Examples.**
  - [internal/services/vm/helpers.go](internal/services/vm/helpers.go) — derived
    sets, scalar->list mirroring, create/update body builders.
  - [internal/services/loadbalancer/helpers.go](internal/services/loadbalancer/helpers.go)
    — divergent create vs read types, default health check, single-listener
    extraction.
  - [internal/services/securitygrouprule/helpers.go](internal/services/securitygrouprule/helpers.go)
    — tiny derived-id and first-element getters that defer to `core`.
- **Rules.**
  - Every `mapping_overrides` entry must carry a `reason` — `make audit` fails
    otherwise.
  - Helpers are hand-written, so they do **not** count as scaffold drift; the
    generator only diffs the files it produces.
  - Keep helper names referenced **verbatim** in the descriptor so regeneration
    stays stable.

## Where to look

- [docs/internal/ADDING_A_FEATURE.md](docs/internal/ADDING_A_FEATURE.md) — end-to-end feature workflow.
- [docs/internal/DESCRIPTOR.md](docs/internal/DESCRIPTOR.md) — every descriptor field.
- [docs/internal/COMMANDS.md](docs/internal/COMMANDS.md) — all `make` targets and `cmd/scaffold` subcommands.
- [docs/internal/RESOURCE_REGISTRY.md](docs/internal/RESOURCE_REGISTRY.md) — async/retry/owner reference table.
- [docs/internal/SCAFFOLD_GRADUATION_LOG.md](docs/internal/SCAFFOLD_GRADUATION_LOG.md) — recipes used to make hard resources fully generated.
- `cmd/scaffold/` — the generator (`type_mapper.go` is the field mapper).
- `internal/core/` — API-call layer; `internal/utils/` — shared TF<->SDK converters.

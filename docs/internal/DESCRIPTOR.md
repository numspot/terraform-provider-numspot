# Descriptor reference (`descriptor.yaml`)

Every generated resource/data source is described by one
`internal/services/<name>/descriptor.yaml`. It is the **single source of truth**
consumed by three tools:

| Tool | Reads the descriptor for |
|---|---|
| `cmd/scaffold` (`gen`/`audit`) | CRUD skeleton, field mapping, core-signature inference |
| `cmd/inject-schema-overrides` | `*_spec.json` post-processing (sets, optionality, plan modifiers, synthetic attrs, …) |
| `cmd/scaffold drift` | `scaffold_check` / `spec_check` drift gates |

This file documents every field. It is kept in lockstep with the Go structs in
[cmd/scaffold/descriptor.go](../../cmd/scaffold/descriptor.go) and
[cmd/inject-schema-overrides/main.go](../../cmd/inject-schema-overrides/main.go),
and with the machine-readable [descriptor.schema.json](descriptor.schema.json).
For the end-to-end workflow (when to set each field, worked examples) see
[ADDING_A_FEATURE.md](ADDING_A_FEATURE.md); for commands see [COMMANDS.md](COMMANDS.md).

## Editor autocomplete and validation

`descriptor.schema.json` is a JSON Schema (draft 2020-12). Point your editor at it
to get completion + inline validation:

- **VS Code / any `yaml-language-server`**: `make scaffold-init` already writes a
  modeline as the first line of new descriptors:

  ```yaml
  # yaml-language-server: $schema=../../../docs/internal/descriptor.schema.json
  ```

  Add that line to an existing descriptor to opt it in.
- **JetBrains (IDEA/GoLand)**: either the modeline above is honoured, or map it
  under *Settings → Languages & Frameworks → Schemas and DTDs → JSON Schema
  Mappings* (schema file `docs/internal/descriptor.schema.json`, file pattern
  `internal/services/*/descriptor.yaml`).

The schema is validated in CI indirectly: any field it rejects is also rejected
by `cmd/scaffold` / `cmd/inject-schema-overrides`.

---

## Identity and naming

| Field | Type | Required | Meaning |
|---|---|:---:|---|
| `name` | string (snake_case) | yes | TF resource name. The Go package folder is the concatenated lowercase form (underscores removed): `dhcp_options` → `dhcpoptions`. |
| `go_name` | string (PascalCase) | yes | Go identifier for constructors and the core function prefix (e.g. `VPC`, `OrganisationRole`). |
| `singular` | string | yes | TF model prefix (`<Singular>Model`) and SDK singular type lookup (case/separator-insensitive). |
| `plural` | string | yes | TF plural and SDK plural lookup (e.g. `ReadOrganisationRolesParams`). |
| `sdk_singular` / `sdk_plural` | string | no | SDK type override (`api.<X>`). Only for genuine name differences (e.g. `DhcpOptionsSet`) or to disambiguate an ambiguous normalized match — **not** for casing. |
| `package` | string | no | **Deprecated.** Redundant with the derived folder; `cmd/scaffold audit` warns if present. |

## Behavior flags

| Field | Type | Default | `true` emits |
|---|---|:---:|---|
| `scope` | `space` \| `organisation` | — (required) | `provider.SpaceID` vs `provider.OrganisationID`. |
| `has_tags` | bool | `false` | Tags in serialize/items, `<GoName>Tags()`, `Update<Resource>Tags()`, and IR `tags` normalization. |
| `is_async` | bool | `false` | `pendingStates`/`targetStates`, `RetryRead<Resource>()`, Create polls until ready. |
| `has_update` | bool | `false` | `Update()` with tags diff + any `core.update.calls`. |
| `create_direct` | bool | `true` | Create returns `JSON201` directly; `false` → Create calls `RetryRead` after POST. |
| `has_resource` | bool | `true` | `resource_<name>.go` + core Create/Read/Delete. |
| `has_datasource` | bool | `true` | `datasource_<name>s.go` + core ReadList. |
| `create_response_type` | string | — | SDK type returned by Create when it differs from `singular`. |
| `no_import` | bool | `false` | Override `ImportState` with an explicit error (for resources keyed on something other than `Id`). |

## Drift gates

| Field | Type | Default | Meaning |
|---|---|:---:|---|
| `scaffold_check` | bool | `false` | `cmd/scaffold drift` enforces the generated `.go` matches the descriptor **and** `cmd/scaffold audit` fails on any residual `TODO`. |
| `spec_check` | bool | `false` | `cmd/scaffold drift` regenerates `<name>_spec.json` through the full pipeline and fails on byte-level drift. Independent of `scaffold_check`. |
| `reason` | string | — | **Required** when `scaffold_check` is not `true`: why the resource is not fully generated. |

## Pattern flags (preferred mapping knobs)

Prefer these high-level flags over `mapping_overrides` — they keep the descriptor
readable and are validated by `cmd/scaffold audit`.

| Field | Type | Meaning |
|---|---|---|
| `id_source` | string | TF `id` is stored in this SDK field (e.g. `Name`). Replaces per-kind `Id -> Name` overrides. |
| `preserve_from_state` | []string | Fields not returned by the API but kept in TF state (skipped when serializing read/create responses). |
| `exclude_from_request` | []string | Fields not sent in the create/update request body. |
| `computed_fields` | []string | Fields returned by the API but not user-provided (skipped in the create request, still mapped on read). |
| `managed_by_other_resource` | []string | Fields owned by a dedicated TF resource (e.g. SG rules). Like `exclude_from_request`, also checked by drift. |
| `null_on_read` | []string | Nested leaf paths (dotted snake_case) present in the TF schema but absent from the read SDK struct: read serializer emits a typed null, read deserializer skips them. |
| `constants` | map[string]string | TF field → constant Go expression injected into the mapping. |

## `mapping_overrides` (escape hatch)

Each override **requires a `reason`** (audit fails otherwise). Five buckets, each a
list of override entries:

| Bucket | Direction |
|---|---|
| `serialize` | read response: TF model → `api.<Singular>` |
| `create_response` | create response: TF model → `api.<create_response_type>` |
| `deserialize` | request body: TF model → `api.Create<Singular>JSONRequestBody` |
| `datasource` | datasource items: `api.<Singular>` → `ItemsValue` |
| `datasource_params` | datasource read filters: `<Model>` → `api.<ReadListParamsType>` |

Override entry fields:

| Field | Type | Meaning |
|---|---|---|
| `tf_field` | string | TF model field name (PascalCase). |
| `path` | string | Resource-side dotted path that scopes the override to a specific top-level or nested field (e.g. `nics.private_ips`). When set, the override only applies at that path and takes precedence over a name-based match. Use this when a top-level field and a nested leaf share the same name but need different expressions. |
| `sdk_field` | string | Force a differently-named SDK field. |
| `expression` | string | Replace the generated expression verbatim (advanced). |
| `skip` | bool | Omit the field from the generated mapping. |
| `reason` | string | Required: why the exception exists. |

`nested_type_overrides` (`tf_field` → `value_type`) overrides the generated TF
Value type for a nested object (default `<TF field>Value`).

## `core` (core-signature inference)

If `internal/core/<name>.go` exists, `cmd/scaffold` parses it for the real
signatures; the `core.*` values below are used only when the parser cannot infer
something (so hand-written orchestration is preserved by default).

| Path | Type | Meaning |
|---|---|---|
| `core.key_expr` | string | Resource key for Read/Delete + tag-update (default `state.Id.ValueString()`). Name-keyed: `state.Name.ValueString()`. |
| `core.regenerate` | bool | Rewrite `internal/core/<name>.go` (default: preserve). |
| `core.create.extra_args` | []string | Extra `Create` args (literal Go expressions) after `(ctx, provider, body)`. |
| `core.create.request_body_type` | string | Override the SDK create body type name when it isn't `Create<Singular>JSONRequestBody`. |
| `core.create.no_request_body` | bool | Create takes no JSON body; no deserializer emitted, Create passes only `extra_args` + tags. |
| `core.read.function_name` | string | Override the singular read function name (e.g. `ReadDHCPOption`). |
| `core.read_list.function_prefix` / `function_suffix` | string | Override the core list name parts (e.g. suffix `WithParams`). |
| `core.read_list.params_type` | string | Type passed to the list function (e.g. `ReadVpcsParams`). |
| `core.read_list.function_name` | string | Override the list function name outright (when uninferrable). |
| `core.delete.extra_args` | []string | Extra `Delete` args. |
| `core.update.calls[]` | list | Non-tag attribute updates: `function_name`, `when_changed` (≥1 field), `args` (literal Go exprs). Fires when any `when_changed` field differs. |
| `core.update.refresh_only` | bool | Update performs no API mutation (re-read + serialize + copy `preserve_from_state` from plan). Mutually exclusive with `has_update`/`calls`. |

`parent_key` models a parent-scoped sub-resource keyed on `(parent, id)`:

| Field | Type | Meaning |
|---|---|---|
| `parent_key.field` | string | Go model field carrying the parent id (e.g. `ClusterId`). |
| `parent_key.tf_name` | string | snake_case name used in the composite-import error (defaults to snake_case of `field`). |
| `parent_key.import` | bool | Emit composite `"<parent>/<id>"` `ImportState`. |

## Spec-injection knobs (`cmd/inject-schema-overrides`)

These re-apply, on every `make generate-spec`, the parts `tfplugingen-openapi`
cannot derive — so the committed `*_spec.json` stays reproducible. Paths are
dotted snake_case (segments descend nested objects); an `items.`-prefixed path
targets the datasource leaf.

| Field | Type | Meaning |
|---|---|---|
| `requires_replace` | []string | Top-level attributes that force replacement (sugar for a `plan_modifiers` `requires_replace`). |
| `plan_modifiers[]` | list | `path` + `kind` (`requires_replace`, `requires_replace_if_configured`, `use_state_for_unknown`, `requires_replace_if`); the predicate form adds `func` + `description`/`description_markdown`. |
| `validators[]` | list | Non-OAS validators: `path`, `schema_definition` (full Go expr), `imports`. Do **not** list OAS-derived validators. |
| `defaults[]` | list | Non-OAS static defaults: `path`, `static` (typed against the attribute). |
| `clear_defaults` | []string | Inverse of `defaults`: drop the OAS-emitted default on these paths. |
| `clear_validators` | []string | Drop all validators on these paths. |
| `descriptions[]` | list | Override an OAS-derived attribute's `description` (`path` + `value`) — the only place to reword an OAS attribute reproducibly. |
| `sets` | []string | Convert `list`/`list_nested` to `set`/`set_nested` (order-insensitive collections). `tags` is automatic — don't list it. |
| `optionality[]` | list | Override `computed_optional_required` (`path` + `value`: `computed`/`computed_optional`/`optional`/`required`). |
| `renames[]` | list | Rename a colliding nested leaf (`path` → `to`), applied resource-side and at `items.<path>`. Leaf type unchanged. |
| `synthetic_attributes[]` | list | Provider-only attrs absent from the OAS (`path`, `type`, `value`, `element_type` for lists, `description`, `deprecation_message`, recursive `attributes`). Re-added resource-side before all other passes. |

> Descriptions otherwise come from the OAS: there is deliberately no general
> `descriptions:` knob beyond the reproducibility cases above. Fix wording
> upstream in the OAS when possible (see ADDING_A_FEATURE.md).

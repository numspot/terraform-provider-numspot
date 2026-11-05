# Scaffold Graduation Log

Recipes used to make resources fully generated (`scaffold_check: true`). Each
entry documents the manual logic that blocked graduation and the descriptor /
helpers mechanism that expresses it. See [ADDING_A_FEATURE.md](ADDING_A_FEATURE.md)
for the general workflow and [DESCRIPTOR.md](DESCRIPTOR.md) for descriptor
fields.

## image — nested-object deserialization via `helpers.go`

- **Problem.** The create flow maps a nested object (`block_device_mappings[].bsu`)
  into the request body, but the generator's nested deserialize branch emits
  `obj.As(ctx, &e, ...)` targeting the generated `attr.Value` struct (e.g.
  `BsuValue`). The framework's reflection layer rejects custom `attr.Value`
  structs as `obj.As` targets, so the generated branch cannot compile — the
  resource carried hand-written `deserializeBsu` /
  `deserializeBlockDeviceMappings` and ran with `scaffold_check: false`.
- **Recipe.**
  1. Move the manual mapping functions verbatim into a hand-written
     `internal/services/<name>/helpers.go` (never overwritten by
     `make scaffold`). Create the helpers **before** regenerating.
  2. Reference the root field from the descriptor:
     ```yaml
     mapping_overrides:
       deserialize:
         - tf_field: BlockDeviceMappings
           expression: utils.PointerOf(deserializeBlockDeviceMappings(ctx, tf.BlockDeviceMappings))
           reason: obj.As cannot target the generated attr.Value structs; nested bsu deserialized manually in helpers.go
     ```
     An `expression` override short-circuits the generator's nested deserialize
     construction, so the broken `obj.As` branch is never emitted; the serialize
     side stays generated.
  3. Set `scaffold_check: true` and delete the top-level `reason:` (audit
     rejects a reason while the check is enabled). Every override needs its own
     `reason` — `make audit` enforces both.
  4. Run `make scaffold RESOURCE_NAME=<name>`, then the verification chain
     (`make fmt && go build ./... && go vet ./... && make audit &&
     make check-scaffold-drift`).
- **Caveat.** After regenerating, review the diff of the remaining scalar
  conversions: the canonical type pairs deserialize with raw
  `ValueStringPointer()` / `ValueBoolPointer()`, which send `""`/`false` for
  **unknown** (unset computed_optional) plan values instead of omitting the
  field. Keep the unknown-safe `utils.FromTfStringToStringPtr` /
  `utils.FromTfBoolToBoolPtr` conversions via one `mapping_overrides.deserialize`
  entry per field (see image and volume descriptors for the exact wording).
- **Example.** `internal/services/image/` — `helpers.go` +
  `descriptor.yaml` (`mapping_overrides.deserialize`), unit-tested in
  `helpers_test.go` (the only code in the pipeline the drift guard does not
  cover, since helpers are hand-written by design).

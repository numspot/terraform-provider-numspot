package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// errDescriptorNameRequired is returned when a descriptor has no name.
	errDescriptorNameRequired = errors.New("descriptor: name is required")
	// errDescriptorGoNameRequired is returned when a descriptor has no go_name.
	errDescriptorGoNameRequired = errors.New("descriptor: go_name is required")
	// errDescriptorSingularRequired is returned when a descriptor has no singular.
	errDescriptorSingularRequired = errors.New("descriptor: singular is required")
	// errDescriptorPluralRequired is returned when a descriptor has no plural.
	errDescriptorPluralRequired = errors.New("descriptor: plural is required")
	// errDescriptorScopeInvalid is returned when a descriptor has an invalid scope.
	errDescriptorScopeInvalid = errors.New("descriptor: scope must be 'space' or 'organisation'")
	// errDescriptorNoResourceOrDatasource is returned when a descriptor defines
	// neither a resource nor a data source.
	errDescriptorNoResourceOrDatasource = errors.New("descriptor: at least one of has_resource or has_datasource must be true")
	// errDescriptorUpdateCallFunctionName is returned when an update call lacks a
	// function name.
	errDescriptorUpdateCallFunctionName = errors.New("descriptor: core.update.calls.function_name is required")
	// errDescriptorUpdateCallWhenChanged is returned when an update call lacks
	// fields to watch.
	errDescriptorUpdateCallWhenChanged = errors.New("descriptor: core.update.calls.when_changed must list at least one field")
	// errDescriptorUpdateCallsRequireUpdate is returned when update calls are
	// defined without has_update.
	errDescriptorUpdateCallsRequireUpdate = errors.New("descriptor: core.update.calls requires has_update: true")
	// errDescriptorUpdateCallUnlessChanged is returned when an update call's
	// unless_changed field is not a when_changed field of a sibling call, so
	// the mutual-exclusion partition between update paths can silently drift.
	errDescriptorUpdateCallUnlessChanged = errors.New(
		"descriptor: core.update.calls.unless_changed field must appear in a sibling call's when_changed",
	)
	// errDescriptorUpdateCallFieldOverlap is returned when an update call lists
	// the same field in both when_changed and unless_changed, a dead condition
	// (plannedChange(x) && !plannedChange(x)) that can never fire.
	errDescriptorUpdateCallFieldOverlap = errors.New(
		"descriptor: core.update.calls field must not appear in both when_changed and unless_changed of the same call",
	)
	// errDescriptorReadIfUnchangedWithoutUpdate is returned when
	// read_if_unchanged is set but no update call and no tag reconciliation can
	// populate the API response, leaving the generated Update to serialize a
	// nil response.
	errDescriptorReadIfUnchangedWithoutUpdate = errors.New("descriptor: core.update.read_if_unchanged requires core.update.calls or has_tags")
	// errDescriptorRefreshOnlyConflict is returned when refresh_only conflicts
	// with update configuration.
	errDescriptorRefreshOnlyConflict = errors.New("descriptor: core.update.refresh_only is mutually exclusive with has_update / core.update.calls")
)

type OverrideConfig struct {
	TFField    string `yaml:"tf_field"`
	Path       string `yaml:"path"`
	SDKField   string `yaml:"sdk_field"`
	Expression string `yaml:"expression"`
	Skip       bool   `yaml:"skip"`
	Reason     string `yaml:"reason"`
}

type NestedTypeOverride struct {
	TFField   string `yaml:"tf_field"`
	ValueType string `yaml:"value_type"`
}

type MappingOverrides struct {
	Serialize      []OverrideConfig `yaml:"serialize"`
	CreateResponse []OverrideConfig `yaml:"create_response"`
	Deserialize    []OverrideConfig `yaml:"deserialize"`
	Datasource     []OverrideConfig `yaml:"datasource"`
	// DatasourceParams overrides the datasource *filter params* builder (distinct
	// from Datasource, which overrides the item serializer). Kept separate because
	// a TF attribute can map differently as a read filter vs a serialized item.
	DatasourceParams []OverrideConfig `yaml:"datasource_params"`
}

type CoreCreateConfig struct {
	ExtraArgs []string `yaml:"extra_args"`
	// RequestBodyType overrides the SDK create request body type name (without the
	// api. prefix). By default the generator derives it from the singular/plural
	// (Create<Singular>JSONRequestBody); set this when the create body uses a
	// type that does not follow that pattern (e.g. postgres_cluster's
	// PostgresClusterCreationRequest, whose JSON alias is the unrelated
	// PostgresqlCreateClusterJSONRequestBody).
	RequestBodyType string `yaml:"request_body_type"`
	// NoRequestBody indicates the SDK create call takes no JSON request body
	// (e.g. internet gateways / public IPs are created empty and configured via
	// follow-up link calls). When true the generator emits no create-body
	// deserializer and the Create call passes only extra_args and tags.
	NoRequestBody bool `yaml:"no_request_body"`
}

type CoreReadConfig struct {
	FunctionName string `yaml:"function_name"`
}

type CoreReadListConfig struct {
	FunctionPrefix string `yaml:"function_prefix"`
	FunctionSuffix string `yaml:"function_suffix"`
	ParamsType     string `yaml:"params_type"`
	// FunctionName overrides the core list function name outright (like
	// CoreReadConfig.FunctionName for the singular read). Needed when the list
	// function cannot be inferred from its signature (e.g. a parent-scoped list
	// taking a string parent id rather than an api.* params struct).
	FunctionName string `yaml:"function_name"`
}

type CoreDeleteConfig struct {
	ExtraArgs []string `yaml:"extra_args"`
}

// CoreUpdateCall is a single non-tag attribute update call emitted in the
// generated Update method. Each call is gated on its WhenChanged fields and
// passes Args as literal Go expressions after (ctx, r.Provider). Tag
// reconciliation (has_tags) is emitted separately; these calls model the
// resource's attribute mutations (e.g. UpdateSubnetAttributes).
type CoreUpdateCall struct {
	// FunctionName is the core function to call (e.g. UpdateSubnetAttributes).
	FunctionName string `yaml:"function_name"`
	// WhenChanged lists TF model field names (e.g. MapPublicIpOnLaunch); the call
	// fires when any of them differs between plan and state. Required.
	WhenChanged []string `yaml:"when_changed"`
	// UnlessChanged lists TF model field names that must *not* differ between
	// plan and state for the call to fire. Use it for mutually-exclusive update
	// paths (e.g. a security-group-only update that must be skipped when other
	// attributes also changed).
	UnlessChanged []string `yaml:"unless_changed"`
	// Args are literal Go expressions passed after (ctx, r.Provider), evaluated in
	// the Update method scope where `plan` and `state` are in scope (e.g.
	// state.Id.ValueString(), plan.MapPublicIpOnLaunch.ValueBool()).
	Args []string `yaml:"args"`
}

type CoreUpdateConfig struct {
	Calls []CoreUpdateCall `yaml:"calls"`
	// RefreshOnly emits an Update method that performs no API mutation: it
	// re-reads the resource (via CoreReadFuncName/CoreKeyExpr), serializes it,
	// and copies the preserve_from_state fields from the *plan* before saving
	// state. Use it for resources whose only mutable attributes are config-only
	// metadata not backed by an update endpoint (e.g. a Terraform `timeouts {}`
	// block): with has_update: false the default no-op Update would drop the
	// edited value and trigger "inconsistent result after apply". RefreshOnly is
	// mutually exclusive with has_update / core.update.calls.
	RefreshOnly bool `yaml:"refresh_only"`
	// UnknownSafe uses a plannedChange() helper for update conditions instead of
	// the raw framework Equal() check. Optional+Computed attributes that are not
	// configured resolve to unknown at plan time; treating unknown as a change
	// would force unnecessary updates (e.g. stopping a VM for an attribute that
	// did not actually change).
	UnknownSafe bool `yaml:"unknown_safe"`
	// ReadIfUnchanged re-reads the resource from the API when none of the update
	// calls or tag reconciliation fired, so the serialized state is never built
	// from a nil response.
	ReadIfUnchanged bool `yaml:"read_if_unchanged"`
}

// PostSerializeHook declares a helper call emitted after the resource is
// serialized from an API response. Used for enrichment that the generated
// serializer cannot express (e.g. reading related APIs to fill write-only
// nested fields).
type PostSerializeHook struct {
	// Helper is the helper function name (must be defined in the resource package).
	Helper string `yaml:"helper"`
	// Args are literal Go expressions passed to the helper. The response
	// diagnostics variable `response.Diagnostics` is in scope.
	Args []string `yaml:"args"`
}

// PostSerializeConfig declares hooks to run after serialization in the Create,
// Read and Update methods.
type PostSerializeConfig struct {
	Create []PostSerializeHook `yaml:"create"`
	Read   []PostSerializeHook `yaml:"read"`
	Update []PostSerializeHook `yaml:"update"`
}

type CoreConfig struct {
	Create   CoreCreateConfig   `yaml:"create"`
	Read     CoreReadConfig     `yaml:"read"`
	ReadList CoreReadListConfig `yaml:"read_list"`
	Delete   CoreDeleteConfig   `yaml:"delete"`
	Update   CoreUpdateConfig   `yaml:"update"`
	// KeyExpr overrides the Go expression used as the resource key passed to the
	// core Read/Delete (and tag-update id) calls. It defaults to
	// state.Id.ValueString(). Name-keyed resources set it to e.g.
	// state.Name.ValueString(); the `state` variable is in scope at all three
	// call sites.
	KeyExpr    string `yaml:"key_expr"`
	Regenerate bool   `yaml:"regenerate"`
}

func (d *Descriptor) CoreFunctionPrefix() string {
	if d.inferredCore != nil && d.inferredCore.CreateFuncName != "" {
		return strings.TrimPrefix(d.inferredCore.CreateFuncName, "Create")
	}
	if d.Core.ReadList.FunctionPrefix != "" {
		return d.Core.ReadList.FunctionPrefix
	}
	return d.GoName
}

// PlanModifierSpec declares a single plan modifier to re-inject into the
// *_spec.json IR by cmd/inject-schema-overrides (tfplugingen-openapi does not emit
// plan modifiers, so they would be lost on every regeneration).
type PlanModifierSpec struct {
	// Path targets the attribute, as dotted snake_case segments matching the
	// *_spec.json schema (e.g. "name" or "network_interfaces.private_ips.private_ip").
	// Each segment after the first descends into the nested object's attributes.
	Path string `yaml:"path"`
	// Kind selects the plan modifier. Supported: "requires_replace",
	// "requires_replace_if_configured", "use_state_for_unknown",
	// "requires_replace_if" (the latter requires Func + Description).
	Kind string `yaml:"kind"`
	// Func is the package-local predicate name for kind=requires_replace_if
	// (e.g. "ReplaceVolumeSize", defined in the resource's package).
	Func string `yaml:"func"`
	// Description (and optional DescriptionMarkdown, defaulting to Description) are
	// the two human-readable args RequiresReplaceIf takes.
	Description         string `yaml:"description"`
	DescriptionMarkdown string `yaml:"description_markdown"`
}

// ValidatorSpec declares a custom validator to re-inject into the *_spec.json IR
// by cmd/inject-schema-overrides. Only validators that tfplugingen-openapi does
// NOT derive from the OAS (e.g. stringvalidator.ConflictsWith) need declaring;
// OAS-derived validators (LengthAtMost, OneOf, Between, ...) are regenerated and
// must not be listed here. Injected validators are appended to whatever the OAS
// produced for the attribute (deduplicated by SchemaDefinition).
type ValidatorSpec struct {
	// Path targets the attribute, as dotted snake_case segments (see PlanModifierSpec).
	Path string `yaml:"path"`
	// SchemaDefinition is the Go expression, e.g.
	// `stringvalidator.ConflictsWith(path.MatchRoot("vm_id"))`.
	SchemaDefinition string `yaml:"schema_definition"`
	// Imports are the package paths the SchemaDefinition needs (e.g. the
	// stringvalidator and path packages).
	Imports []string `yaml:"imports"`
}

// DefaultSpec declares a static default to re-inject into the *_spec.json IR by
// cmd/inject-schema-overrides. Only defaults that tfplugingen-openapi does NOT
// emit from the OAS need declaring. The Static value is typed against the target
// attribute (bool/int64/string/...).
type DefaultSpec struct {
	// Path targets the attribute, as dotted snake_case segments (see PlanModifierSpec).
	Path string `yaml:"path"`
	// Static is the literal default value (bool, integer or string), matched to
	// the attribute's type.
	Static any `yaml:"static"`
}

//nolint:gocritic // keep Descriptor near the type aliases it uses in doc comments
type Descriptor struct {
	Name                string               `yaml:"name"`
	Package             string               `yaml:"package"`
	GoName              string               `yaml:"go_name"`
	Singular            string               `yaml:"singular"`
	Plural              string               `yaml:"plural"`
	SDKSingular         string               `yaml:"sdk_singular"`
	SDKPlural           string               `yaml:"sdk_plural"`
	Scope               string               `yaml:"scope"`
	HasTags             bool                 `yaml:"has_tags"`
	IsAsync             bool                 `yaml:"is_async"`
	HasUpdate           bool                 `yaml:"has_update"`
	CreateDirect        bool                 `yaml:"create_direct"`
	HasResource         bool                 `yaml:"has_resource"`
	HasDatasource       bool                 `yaml:"has_datasource"`
	CreateResponseType  string               `yaml:"create_response_type"`
	MappingOverrides    MappingOverrides     `yaml:"mapping_overrides"`
	NestedTypeOverrides []NestedTypeOverride `yaml:"nested_type_overrides"`
	Core                CoreConfig           `yaml:"core"`
	ScaffoldCheck       bool                 `yaml:"scaffold_check"`
	Reason              string               `yaml:"reason"`

	// NoImport suppresses Terraform import for the generated resource. When true
	// the generated resource overrides ImportState to return an explicit error
	// instead of inheriting BaseResource's id-passthrough (which is meaningless
	// for resources keyed on something other than Id).
	NoImport bool `yaml:"no_import"`

	IDSource          string   `yaml:"id_source"`
	PreserveFromState []string `yaml:"preserve_from_state"`
	// RequiresReplace lists schema attribute names (snake_case, as they appear in
	// *_spec.json) that must force resource replacement when changed. It is sugar
	// for a PlanModifiers entry with kind=requires_replace targeting a top-level
	// attribute. It is not used by the scaffold itself; cmd/inject-schema-overrides
	// reads it (and PlanModifiers) to re-apply plan modifiers to the generated IR
	// after `make generate-spec` (tfplugingen-openapi does not emit plan
	// modifiers). Declared here so the descriptor stays the single source of truth.
	RequiresReplace []string `yaml:"requires_replace"`
	// PlanModifiers declares the full set of plan modifiers to re-inject into the
	// *_spec.json IR after generation. Each entry targets an attribute by dotted
	// Path (snake_case segments, descending nested objects) and a Kind. See
	// cmd/inject-schema-overrides for the supported kinds. Like RequiresReplace,
	// this is consumed only by that post-processor, not by the scaffold itself.
	PlanModifiers []PlanModifierSpec `yaml:"plan_modifiers"`
	// Validators and Defaults declare non-OAS-derived custom validators and static
	// defaults to re-inject into the IR (also via cmd/inject-schema-overrides).
	Validators             []ValidatorSpec   `yaml:"validators"`
	Defaults               []DefaultSpec     `yaml:"defaults"`
	ExcludeFromRequest     []string          `yaml:"exclude_from_request"`
	ManagedByOtherResource []string          `yaml:"managed_by_other_resource"`
	Constants              map[string]string `yaml:"constants"`
	ComputedFields         []string          `yaml:"computed_fields"`
	// Renames mirrors the cmd/inject-schema-overrides knob that renames a nested
	// leaf in the *_spec.json IR to avoid duplicate generated Go types (e.g. two
	// nested objects both named `link_public_ip`). The scaffold itself does not
	// rename anything; it only needs the mapping so nested field matching can
	// recover the original SDK leaf name (the SDK struct still uses the OAS name).
	Renames []RenameSpec `yaml:"renames"`

	// NullOnRead lists nested leaf paths (dotted snake_case tfsdk tags, e.g.
	// `block_device_mappings.bsu.iops`) that exist in the Terraform schema but are
	// absent from the *read* SDK struct. The read serializer emits a typed null
	// for them (instead of a TODO), and the read struct's deserializer skips them.
	// Use this for "superset" TF nested types whose write-only leaves the read
	// payload does not echo back (the create/update body is mapped separately).
	NullOnRead []string `yaml:"null_on_read"`

	// NullIfUnknown lists nested object paths (dotted snake_case tfsdk tags) whose
	// deserialize helper must treat an unknown plan value as null. Optional+Computed
	// nested attributes that are not configured surface as UNKNOWN (not null) at
	// plan time; sending a zeroed object would fail API validation, so the helper
	// returns nil for unknown as well as null.
	NullIfUnknown []string `yaml:"null_if_unknown"`

	// PostSerialize declares helper calls emitted after the resource is serialized
	// from an API response in Create, Read and Update.
	PostSerialize *PostSerializeConfig `yaml:"post_serialize"`

	// ParentKey models a parent-scoped sub-resource whose CRUD is keyed on a
	// parent id in addition to its own id (e.g. a kubernetes_nodepool keyed on
	// (cluster_id, id)). When set, the generated Read/Delete calls and the
	// datasource list call thread the parent id as an extra leading argument, and
	// (when Import is true) ImportState parses a composite "<parent>/<id>" id.
	// The Create call passes the parent id via core.create.extra_args (the parent
	// id is the last create argument in the conformed core signature).
	ParentKey *ParentKeyConfig `yaml:"parent_key"`

	inferredCore *CoreSignatures
}

// ParentKeyConfig configures a parent-scoped composite key (see
// Descriptor.ParentKey). The conformed core functions take the parent id as a
// string and parse it internally, so the generated arguments are plain
// plan/state field reads.
type ParentKeyConfig struct {
	// Field is the Go model field carrying the parent id (e.g. ClusterId). The
	// generated code reads plan.<Field>/state.<Field>.
	Field string `yaml:"field"`
	// TFName is the snake_case attribute name for the parent id, used only in the
	// composite-import error message (defaults to the snake_case of Field).
	TFName string `yaml:"tf_name"`
	// Import emits a composite "<parent>/<id>" ImportState that sets <Field> and
	// Id from the two id segments.
	Import bool `yaml:"import"`
}

// RenameSpec maps a renamed schema leaf (To, the final attribute name) back to
// its original OAS path (Path, dotted snake_case). The SDK field corresponds to
// the original leaf (the last Path segment), not the renamed name.
type RenameSpec struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

// OriginalLeafForRenamed returns the original (pre-rename) snake_case leaf for a
// renamed attribute tag, so nested field matching can find the SDK field that
// still carries the OAS name. Returns ("", false) when the tag was not renamed.
func (d *Descriptor) OriginalLeafForRenamed(tag string) (string, bool) {
	for _, r := range d.Renames {
		if r.To != tag {
			continue
		}
		leaf := r.Path
		if i := strings.LastIndex(leaf, "."); i >= 0 {
			leaf = leaf[i+1:]
		}
		if leaf != "" {
			return leaf, true
		}
	}
	return "", false
}

func (d *Descriptor) NestedTypeName(tfFieldName string) string {
	for _, o := range d.NestedTypeOverrides {
		if o.TFField == tfFieldName {
			return o.ValueType
		}
	}
	return tfFieldName + "Value"
}

func (d *Descriptor) IsIDSource(tfFieldName string) bool {
	return strings.EqualFold(tfFieldName, "Id") && d.IDSource != ""
}

func (d *Descriptor) IsPreserveFromState(tfFieldName string) bool {
	for _, f := range d.PreserveFromState {
		if f == tfFieldName {
			return true
		}
	}
	return false
}

func (d *Descriptor) IsExcludeFromRequest(tfFieldName string) bool {
	if d.IsManagedByOtherResource(tfFieldName) {
		return true
	}
	for _, f := range d.ExcludeFromRequest {
		if f == tfFieldName {
			return true
		}
	}
	return false
}

func (d *Descriptor) IsManagedByOtherResource(tfFieldName string) bool {
	for _, f := range d.ManagedByOtherResource {
		if f == tfFieldName {
			return true
		}
	}
	return false
}

// IsNullOnRead reports whether the dotted snake_case path (e.g.
// "block_device_mappings.bsu.iops") is declared as absent from the read SDK
// struct, so the read serializer should emit a typed null and the read
// deserializer should skip it.
func (d *Descriptor) IsNullOnRead(path string) bool {
	for _, p := range d.NullOnRead {
		if p == path {
			return true
		}
	}
	return false
}

// IsNullIfUnknown reports whether the dotted snake_case path (e.g.
// "autoscaling") is declared as returning nil when the plan value is unknown.
func (d *Descriptor) IsNullIfUnknown(path string) bool {
	for _, p := range d.NullIfUnknown {
		if p == path {
			return true
		}
	}
	return false
}

func (d *Descriptor) IsComputedField(tfFieldName string) bool {
	for _, f := range d.ComputedFields {
		if f == tfFieldName {
			return true
		}
	}
	return false
}

func (d *Descriptor) ConstantValue(tfFieldName string) (string, bool) {
	v, ok := d.Constants[tfFieldName]
	return v, ok
}

func (d *Descriptor) CoreReadListFuncName() string {
	if d.Core.ReadList.FunctionName != "" {
		return d.Core.ReadList.FunctionName
	}
	if d.inferredCore != nil && d.inferredCore.ReadListFuncName != "" {
		return d.inferredCore.ReadListFuncName
	}
	suffix := d.Core.ReadList.FunctionSuffix
	return "Read" + d.Plural + suffix
}

// HasParentKey reports whether this resource is a parent-scoped sub-resource.
func (d *Descriptor) HasParentKey() bool {
	return d.ParentKey != nil && d.ParentKey.Field != ""
}

// ParentKeyField is the Go model field carrying the parent id (e.g. ClusterId).
func (d *Descriptor) ParentKeyField() string {
	if d.ParentKey == nil {
		return ""
	}
	return d.ParentKey.Field
}

// ParentKeyTFName is the snake_case parent attribute name used in the composite
// import error message (defaults to the snake_case of the model field).
func (d *Descriptor) ParentKeyTFName() string {
	if d.ParentKey == nil {
		return ""
	}
	if d.ParentKey.TFName != "" {
		return d.ParentKey.TFName
	}
	return pascalToSnake(d.ParentKey.Field)
}

// ParentKeyRWArg is the Go expression for the parent id at the Read/Delete call
// sites, where `state` is in scope.
func (d *Descriptor) ParentKeyRWArg() string {
	return "state." + d.ParentKeyField() + ".ValueString()"
}

// ParentKeyDatasourceArg is the Go expression for the parent id at the
// datasource list call site, where `plan` is in scope.
func (d *Descriptor) ParentKeyDatasourceArg() string {
	return "plan." + d.ParentKeyField() + ".ValueString()"
}

// ParentKeyImport reports whether to emit a composite "<parent>/<id>"
// ImportState.
func (d *Descriptor) ParentKeyImport() bool {
	return d.HasParentKey() && d.ParentKey.Import
}

func (d *Descriptor) CoreReadListParamsType() string {
	if d.Core.ReadList.ParamsType != "" {
		return d.Core.ReadList.ParamsType
	}
	if d.inferredCore != nil {
		return d.inferredCore.ReadListParamsType
	}
	return "Read" + d.Plural + "Params"
}

func (d *Descriptor) CreateResponseTypeName() string {
	if d.CreateResponseType != "" {
		return d.CreateResponseType
	}
	return ""
}

func (d *Descriptor) ScopeID() string {
	if d.Scope == "organisation" {
		return "OrganisationID"
	}
	return "SpaceID"
}

func (d *Descriptor) ScopeRef() string {
	return "provider." + d.ScopeID()
}

func (d *Descriptor) TFName() string {
	return d.Name
}

func (d *Descriptor) ResourceStructName() string {
	return strings.ToLower(d.Singular[:1]) + d.Singular[1:] + "Resource"
}

func (d *Descriptor) DatasourceStructName() string {
	pkg := d.PackageName()
	// Avoid doubling the "s" when the package name is already plural
	// (dhcpoptions -> dhcpoptionsDataSource, not dhcpoptionssDataSource).
	if strings.HasSuffix(pkg, "s") {
		return pkg + "DataSource"
	}
	return pkg + "sDataSource"
}

func (d *Descriptor) ResourceConstructor() string {
	return "New" + d.GoName + "Resource"
}

// DatasourceConstructor derives the datasource constructor from the plural noun
// rather than GoName+"s". This keeps the usual New<GoName>sDataSource for regular
// resources (where Plural == GoName+"s") while supporting already-plural names
// like DHCPOptions (-> NewDHCPOptionsDataSource, not NewDHCPOptionssDataSource).
func (d *Descriptor) DatasourceConstructor() string {
	return "New" + d.Plural + "DataSource"
}

// TFDatasourceName is the Terraform type name for the (list) datasource. It is
// the resource name pluralized with an "s", but is left untouched when the name
// is already plural (ends in "s"), e.g. dhcp_options -> numspot_dhcp_options.
func (d *Descriptor) TFDatasourceName() string {
	if strings.HasSuffix(d.Name, "s") {
		return d.Name
	}
	return d.Name + "s"
}

// CoreReadFuncName is the singular read function name in internal/core. It
// honours an explicit core.read.function_name override (e.g. ReadDHCPOption,
// which is not Read+prefix) and otherwise defaults to Read<CoreFunctionPrefix>.
func (d *Descriptor) CoreReadFuncName() string {
	if d.Core.Read.FunctionName != "" {
		return d.Core.Read.FunctionName
	}
	return "Read" + d.CoreFunctionPrefix()
}

// CoreKeyExpr is the Go expression used as the resource key for the generated
// core Read/Delete and tag-update calls. It honours core.key_expr and otherwise
// defaults to the resource Id from prior state.
func (d *Descriptor) CoreKeyExpr() string {
	if d.Core.KeyExpr != "" {
		return d.Core.KeyExpr
	}
	return "state.Id.ValueString()"
}

func (d *Descriptor) PackageName() string {
	if d.Package != "" {
		return d.Package
	}
	return strings.ReplaceAll(d.Name, "_", "")
}

func (d *Descriptor) ResourceSubPackage() string {
	return "resource_" + singularSnake(d.Singular)
}

func (d *Descriptor) DatasourceSubPackage() string {
	return d.DatasourceModelPackage()
}

func (d *Descriptor) ResourceModelPackage() string {
	return "resource_" + singularSnake(d.Singular)
}

func (d *Descriptor) DatasourceModelPackage() string {
	return "datasource_" + singularSnake(d.Singular)
}

func singularSnake(s string) string {
	return pascalToSnake(s)
}

func pascalToSnake(s string) string {
	if s == "" {
		return ""
	}
	var out []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '_')
		}
		out = append(out, r)
	}
	return strings.ToLower(string(out))
}

func (d *Descriptor) LocalVar() string {
	pkg := d.PackageName()
	if pkg == "" {
		return ""
	}
	return pkg
}

func (d *Descriptor) SDKSingularName() string {
	if d.SDKSingular != "" {
		return d.SDKSingular
	}
	return d.Singular
}

func (d *Descriptor) SDKPluralName() string {
	if d.SDKPlural != "" {
		return d.SDKPlural
	}
	return d.Plural
}

func (d *Descriptor) Validate() error {
	if err := validateRequiredFields(d); err != nil {
		return err
	}
	if err := validateScope(d); err != nil {
		return err
	}
	if err := validateResourceOrDatasource(d); err != nil {
		return err
	}
	if err := validateUpdateCalls(d); err != nil {
		return err
	}
	if err := validateReadIfUnchanged(d); err != nil {
		return err
	}
	return validateRefreshOnly(d)
}

func validateRequiredFields(d *Descriptor) error {
	if d.Name == "" {
		return errDescriptorNameRequired
	}
	if d.GoName == "" {
		return errDescriptorGoNameRequired
	}
	if d.Singular == "" {
		return errDescriptorSingularRequired
	}
	if d.Plural == "" {
		return errDescriptorPluralRequired
	}
	return nil
}

func validateScope(d *Descriptor) error {
	if d.Scope != "space" && d.Scope != "organisation" {
		return fmt.Errorf("%w, got '%s'", errDescriptorScopeInvalid, d.Scope)
	}
	return nil
}

func validateResourceOrDatasource(d *Descriptor) error {
	if !d.HasResource && !d.HasDatasource {
		return errDescriptorNoResourceOrDatasource
	}
	return nil
}

func validateUpdateCalls(d *Descriptor) error {
	for i, c := range d.Core.Update.Calls {
		if c.FunctionName == "" {
			return fmt.Errorf("%w[%d]", errDescriptorUpdateCallFunctionName, i)
		}
		if len(c.WhenChanged) == 0 {
			return fmt.Errorf("%w[%d]", errDescriptorUpdateCallWhenChanged, i)
		}
		for _, f := range c.UnlessChanged {
			if whenChangedHas(&c, f) {
				return fmt.Errorf("%w[%d]: %s", errDescriptorUpdateCallFieldOverlap, i, f)
			}
			if !siblingWhenChangedHas(d.Core.Update.Calls, i, f) {
				return fmt.Errorf("%w[%d]: %s", errDescriptorUpdateCallUnlessChanged, i, f)
			}
		}
		if !d.HasUpdate {
			return errDescriptorUpdateCallsRequireUpdate
		}
	}
	return nil
}

func whenChangedHas(call *CoreUpdateCall, field string) bool {
	for _, f := range call.WhenChanged {
		if f == field {
			return true
		}
	}
	return false
}

func siblingWhenChangedHas(calls []CoreUpdateCall, self int, field string) bool {
	for i, c := range calls {
		if i == self {
			continue
		}
		for _, f := range c.WhenChanged {
			if f == field {
				return true
			}
		}
	}
	return false
}

func validateRefreshOnly(d *Descriptor) error {
	if d.Core.Update.RefreshOnly && (d.HasUpdate || len(d.Core.Update.Calls) > 0) {
		return errDescriptorRefreshOnlyConflict
	}
	return nil
}

// validateReadIfUnchanged rejects an orphan read_if_unchanged: with no update
// call and no tag reconciliation, nothing can populate the API response
// variable, so the generated Update re-reads the resource and then still
// serializes a nil response.
func validateReadIfUnchanged(d *Descriptor) error {
	if d.Core.Update.ReadIfUnchanged && len(d.Core.Update.Calls) == 0 && !d.HasTags {
		return errDescriptorReadIfUnchangedWithoutUpdate
	}
	return nil
}

func ParseDescriptor(path string) (*Descriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading descriptor: %w", err)
	}

	var d Descriptor
	if err = yaml.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parsing descriptor YAML: %w", err)
	}

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return &d, nil
}

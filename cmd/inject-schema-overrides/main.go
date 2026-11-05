// Command inject-schema-overrides re-applies schema details to a generated
// *_spec.json IR that tfplugingen-openapi does not (and cannot) derive from the
// OAS: plan modifiers, hand-maintained custom validators, and static defaults.
// All three are driven by the resource descriptor, so they survive every
// `make generate-spec` regeneration and the descriptor stays the single source
// of truth.
//
// It runs right after the openapi step (mirroring cmd/inject-ignores) and
// round-trips through the same terraform-plugin-codegen-spec types (and the same
// json.MarshalIndent(_, "", "\t") that tfplugingen-openapi uses), so the only
// changes to the file are the injected overrides — no key reordering. As a side
// effect it normalizes legacy malformed plan-modifier blocks (where `imports`
// sat as a sibling of `custom`) to the canonical shape.
//
// Semantics:
//   - sets:          CONVERT a list/list_nested attribute to set/set_nested.
//     openapi always emits ordered lists, but order-insensitive API collections
//     (ids, ips, rules, tags) must be sets to avoid phantom diffs.
//   - optionality:   SET an attribute's computed_optional_required. openapi
//     derives it from the read schema (often `computed`); declare it when the
//     attribute is actually user-settable.
//   - plan_modifiers: SET (openapi emits none, so replacing is safe).
//   - validators:     APPEND to the OAS-derived validators, deduped by
//     schema_definition (only non-OAS validators like ConflictsWith are declared).
//   - clear_validators: DROP all validators from an attribute (inverse of
//     validators), for OAS enums the provider deliberately does not enforce.
//   - descriptions:   SET an attribute's description, for user-facing fields the
//     provider deliberately rewords away from the OAS text.
//   - defaults:       SET the static default (only non-OAS defaults are declared).
//
// Sets and optionality run first so plan modifiers/validators see the final type
// and the declared optionality is preserved. clear_validators runs after
// validators (and descriptions after optionality) so a path is never both
// declared and undone. The `items.` path prefix targets the matching datasource
// leaf for clear_validators and descriptions, like renames.
//
// Usage: inject-schema-overrides <spec.json> <descriptor.yaml>
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-codegen-spec/code"
	"github.com/hashicorp/terraform-plugin-codegen-spec/datasource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"
	"github.com/hashicorp/terraform-plugin-codegen-spec/spec"
	"gopkg.in/yaml.v3"
)

const datasourceItemsSegment = "items"

var (
	// errNestedAttributeEmptyPath is returned when a nested attribute override
	// has an empty path.
	errNestedAttributeEmptyPath = errors.New("nested attribute has empty path")
	// errMissingType is returned when an override has no type.
	errMissingType = errors.New("missing type")
	// errUnsupportedSyntheticType is returned for an unsupported synthetic type.
	errUnsupportedSyntheticType = errors.New("unsupported synthetic type")
	// errSyntheticListMissingElementType is returned when a synthetic list has no
	// element type.
	errSyntheticListMissingElementType = errors.New("list synthetic attribute requires element_type")
	// errUnsupportedElementType is returned for an unsupported list element type.
	errUnsupportedElementType = errors.New("unsupported element_type")
	// errDatasourceNotFound is returned when a referenced datasource is missing.
	errDatasourceNotFound = errors.New("no datasource to synthesize from")
	// errDatasourceItemsMissing is returned when a datasource has no items
	// attribute.
	errDatasourceItemsMissing = errors.New("datasource has no \"items\" attribute")
	// errDatasourceItemsNotCollection is returned when a datasource items
	// attribute is not a nested list or set.
	errDatasourceItemsNotCollection = errors.New("datasource \"items\" is not a nested list/set")
	// errInvalidOptionality is returned for an invalid optionality value.
	errInvalidOptionality = errors.New("invalid optionality value")
	// errRequiresReplaceIfMissingFields is returned when a requires_replace_if
	// plan modifier lacks required fields.
	errRequiresReplaceIfMissingFields = errors.New("requires_replace_if is missing func or description")
	// errMissingPlanModifierKind is returned when a plan modifier has no kind.
	errMissingPlanModifierKind = errors.New("missing plan modifier kind")
	// errUnknownPlanModifierKind is returned for an unknown plan modifier kind.
	errUnknownPlanModifierKind = errors.New("unknown plan modifier kind")
	// errBoolStaticMismatch is returned when a bool default has the wrong static
	// type.
	errBoolStaticMismatch = errors.New("bool attribute needs a bool static")
	// errStringStaticMismatch is returned when a string default has the wrong
	// static type.
	errStringStaticMismatch = errors.New("string attribute needs a string static")
	// errStaticDefaultUnsupported is returned when a static default is requested
	// for an unsupported attribute type.
	errStaticDefaultUnsupported = errors.New("static default not supported for this attribute type")
	// errIntStaticMismatch is returned when an int64 default has the wrong static
	// type.
	errIntStaticMismatch = errors.New("int64 attribute needs an integer static")
	// errFloatStaticMismatch is returned when a float64 default has the wrong
	// static type.
	errFloatStaticMismatch = errors.New("float64 attribute needs a numeric static")
	// errResourceHasNoSchema is returned when a resource has no schema.
	errResourceHasNoSchema = errors.New("resource has no schema")
	// errResourceNotFound is returned when a resource is not found in the spec.
	errResourceNotFound = errors.New("resource not found")
)

const (
	planModifierBaseImport      = "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	planModifierSubImportPrefix = "github.com/hashicorp/terraform-plugin-framework/resource/schema/"
)

// Supported plan modifier kinds.
const (
	kindRequiresReplace             = "requires_replace"
	kindRequiresReplaceIfConfigured = "requires_replace_if_configured"
	kindUseStateForUnknown          = "use_state_for_unknown"
	kindRequiresReplaceIf           = "requires_replace_if"
)

type planModifierSpec struct {
	Path                string `yaml:"path"`
	Kind                string `yaml:"kind"`
	Func                string `yaml:"func"`
	Description         string `yaml:"description"`
	DescriptionMarkdown string `yaml:"description_markdown"`
}

type validatorSpec struct {
	Path             string   `yaml:"path"`
	SchemaDefinition string   `yaml:"schema_definition"`
	Imports          []string `yaml:"imports"`
}

type defaultSpec struct {
	Path   string `yaml:"path"`
	Static any    `yaml:"static"`
}

type optionalitySpec struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
}

// descriptionSpec overrides an attribute's description. tfplugingen-openapi
// always derives descriptions from the OAS (and for `$ref`-with-enum fields it
// uses the referenced schema's generic description, ignoring the property-level
// one), so a provider that deliberately rewords a user-facing field has no other
// way to keep that wording reproducible. A `items.`-prefixed path targets the
// matching datasource leaf (mirroring renames).
type descriptionSpec struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
}

// renameSpec renames a nested attribute leaf. The OAS reuses one schema name
// (e.g. link_public_ip, security_groups) in several nesting contexts; the
// framework generator then emits one Go type per distinct name and collides.
// Renaming the colliding leaves to unique names (resource-side AND the matching
// datasource items.<path>) lets the framework generate without a clash. Path is
// the resource-side dotted path to the OLD leaf; the datasource is reached at
// "items." + Path. Renames run before every other pass so later passes can
// target the final names. The leaf type is untouched (only the name changes).
type renameSpec struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

// syntheticAttrSpec declares a provider-only attribute that the OAS does not
// define, so the mapper drops it and `make generate-spec` could not reproduce
// it. It is re-added before every other pass so that
// sets/optionality/plan_modifiers/validators/defaults can target it. It is the
// one place a description lives outside the OAS, because the attribute is not in
// the OAS at all. The spec is recursive to cover nested (single_nested) attrs.
// Plain paths target the resource schema; `items.`-prefixed paths target the
// matching datasource `items` element (mirroring renames/descriptions), with
// the read-view optionality (computed) the datasource schema requires.
type syntheticAttrSpec struct {
	Path               string              `yaml:"path"`
	Type               string              `yaml:"type"`
	Value              string              `yaml:"value"`
	ElementType        string              `yaml:"element_type"`
	Description        string              `yaml:"description"`
	DeprecationMessage string              `yaml:"deprecation_message"`
	Attributes         []syntheticAttrSpec `yaml:"attributes"`
}

// coreHead is the minimal slice of the descriptor's `core:` block this tool
// needs. create.no_request_body marks a resource whose SDK create call takes no
// JSON body (e.g. internet gateways / public IPs). tfplugingen-openapi then maps
// no resource schema at all, so we synthesize one from the datasource `items`
// element (see synthesizeResourceFromItems).
type coreCreateHead struct {
	NoRequestBody bool `yaml:"no_request_body"`
}

type coreHead struct {
	Create coreCreateHead `yaml:"create"`
}

type descriptorHead struct {
	Name                string              `yaml:"name"`
	HasTags             bool                `yaml:"has_tags"`
	Core                coreHead            `yaml:"core"`
	SyntheticAttributes []syntheticAttrSpec `yaml:"synthetic_attributes"`
	Renames             []renameSpec        `yaml:"renames"`
	Sets                []string            `yaml:"sets"`
	Optionality         []optionalitySpec   `yaml:"optionality"`
	RequiresReplace     []string            `yaml:"requires_replace"`
	PlanModifiers       []planModifierSpec  `yaml:"plan_modifiers"`
	Validators          []validatorSpec     `yaml:"validators"`
	Defaults            []defaultSpec       `yaml:"defaults"`
	ClearDefaults       []string            `yaml:"clear_defaults"`
	ClearValidators     []string            `yaml:"clear_validators"`
	Descriptions        []descriptionSpec   `yaml:"descriptions"`
}

func main() {
	specPath, descPath, ok := parseArgs()
	if !ok {
		return
	}
	if _, err := os.Stat(descPath); os.IsNotExist(err) {
		return // no descriptor (e.g. legacy resource); nothing to inject
	}
	desc, err := readDescriptor(descPath)
	if err != nil {
		fail("%v", err)
	}

	pmSpecs := mergedPlanModifiers(&desc)
	if !hasDescriptorWork(&desc, pmSpecs) {
		return // nothing declared; leave the spec untouched
	}
	if desc.Name == "" {
		fail("descriptor %s has no name", descPath)
	}

	s, err := loadSpec(specPath)
	if err != nil {
		fail("%v", err)
	}

	res, err := findOrSynthesizeResource(&s, &desc, specPath)
	if err != nil {
		fail("%v", err)
	}

	applyInjections(res, &s, &desc, pmSpecs)

	if err := writeSpec(specPath, s); err != nil {
		fail("%v", err)
	}
}

func parseArgs() (specPath, descPath string, ok bool) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: inject-schema-overrides <spec.json> <descriptor.yaml>")
		return "", "", false
	}
	return os.Args[1], os.Args[2], true
}

func anyPositive(nums ...int) bool {
	for _, n := range nums {
		if n > 0 {
			return true
		}
	}
	return false
}

func hasDescriptorWork(desc *descriptorHead, pmSpecs []planModifierSpec) bool {
	if anyPositive(
		len(pmSpecs),
		len(desc.Validators),
		len(desc.Defaults),
		len(desc.ClearDefaults),
		len(desc.ClearValidators),
		len(desc.Descriptions),
		len(desc.Sets),
		len(desc.Optionality),
		len(desc.SyntheticAttributes),
		len(desc.Renames),
	) {
		return true
	}
	return desc.HasTags || desc.Core.Create.NoRequestBody
}

func loadSpec(path string) (spec.Specification, error) {
	var s spec.Specification
	specData, err := os.ReadFile(path)
	if err != nil {
		return s, fmt.Errorf("reading spec %s: %w", path, err)
	}
	if err := json.Unmarshal(specData, &s); err != nil {
		return s, fmt.Errorf("parsing spec %s: %w", path, err)
	}
	return s, nil
}

func findOrSynthesizeResource(s *spec.Specification, desc *descriptorHead, specPath string) (*resource.Resource, error) {
	for i := range s.Resources {
		if s.Resources[i].Name != desc.Name {
			continue
		}
		res := &s.Resources[i]
		if res.Schema == nil {
			return nil, fmt.Errorf("%w %q in %s", errResourceHasNoSchema, desc.Name, specPath)
		}
		return res, nil
	}
	// create-without-body resources (core.create.no_request_body) get no
	// resource schema from tfplugingen-openapi, so synthesize one from the
	// datasource `items` element before the other passes run.
	if desc.Core.Create.NoRequestBody {
		r, err := synthesizeResourceFromItems(s, desc.Name)
		if err != nil {
			return nil, fmt.Errorf("synthesizing resource %q: %w", desc.Name, err)
		}
		return r, nil
	}
	return nil, fmt.Errorf("%w %q in %s", errResourceNotFound, desc.Name, specPath)
}

func applyInjections(res *resource.Resource, s *spec.Specification, desc *descriptorHead, pmSpecs []planModifierSpec) {
	injectSyntheticAttributes(res, desc.Name, desc.SyntheticAttributes)
	injectDSSyntheticAttributes(s, desc.Name, desc.SyntheticAttributes)
	injectRenames(s, res, desc.Name, desc.Renames)
	if desc.HasTags {
		normalizeTags(res)
	}
	injectSets(res, desc.Name, desc.Sets)
	injectOptionality(res, desc.Name, desc.Optionality)
	injectDescriptions(s, res, desc.Name, desc.Descriptions)
	injectPlanModifiers(res, desc.Name, pmSpecs)
	injectValidators(res, desc.Name, desc.Validators)
	clearValidators(s, res, desc.Name, desc.ClearValidators)
	injectDefaults(res, desc.Name, desc.Defaults)
	clearDefaults(res, desc.Name, desc.ClearDefaults)
}

func writeSpec(path string, s spec.Specification) error {
	out, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return fmt.Errorf("encoding spec: %w", err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("writing spec %s: %w", path, err)
	}
	return nil
}

func readDescriptor(path string) (descriptorHead, error) {
	var d descriptorHead
	data, err := os.ReadFile(path)
	if err != nil {
		return d, fmt.Errorf("reading descriptor %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &d); err != nil {
		return d, fmt.Errorf("parsing descriptor %s: %w", path, err)
	}
	return d, nil
}

// ── tags ────────────────────────────────────────────────────────────

// normalizeTags rewrites the standard top-level `tags` attribute to the
// canonical numspot shape the scaffold template expects: a set_nested,
// computed_optional collection whose key/value are required. openapi derives
// tags from the read schema as a computed list_nested, so every has_tags
// resource needs this normalization. It is a no-op if there is no `tags`
// attribute.
func normalizeTags(res *resource.Resource) {
	attr := resolveAttribute(res.Schema.Attributes, []string{"tags"})
	if attr == nil {
		return
	}
	convertToSet(attr)
	setComputedOptionalRequired(attr, schema.ComputedOptional)
	for _, child := range []string{"key", "value"} {
		if c := resolveAttribute(res.Schema.Attributes, []string{"tags", child}); c != nil {
			setComputedOptionalRequired(c, schema.Required)
		}
	}
}

// ── synthetic attributes (provider-only, absent from the OAS) ────────

// injectSyntheticAttributes adds each declared provider-only attribute to the
// resource schema. Top-level attrs are appended to the schema; dotted paths are
// appended to the parent nested attribute. Insertion order is deterministic, so
// the first regen rewrites the committed *_spec.json and becomes the drift
// baseline (like every other graduation). Runs before all other passes.
// `items.`-prefixed specs target the datasource (see injectDSSyntheticAttributes)
// and are skipped here.
func injectSyntheticAttributes(res *resource.Resource, resName string, specs []syntheticAttrSpec) {
	for _, sp := range specs {
		if sp.Path == "" {
			fail("synthetic attribute with empty path in resource %q", resName)
		}
		if strings.HasPrefix(sp.Path, "items.") {
			continue
		}
		segs := strings.Split(sp.Path, ".")
		leaf := segs[len(segs)-1]
		attr, err := buildSyntheticAttribute(leaf, &sp)
		if err != nil {
			fail("synthetic attribute %q in resource %q: %v", sp.Path, resName, err)
		}
		if resolveAttribute(res.Schema.Attributes, segs) != nil {
			fail("synthetic path %q already exists in resource %q schema", sp.Path, resName)
		}
		if len(segs) == 1 {
			res.Schema.Attributes = append(res.Schema.Attributes, attr)
			continue
		}
		parent := resolveAttribute(res.Schema.Attributes, segs[:len(segs)-1])
		if parent == nil {
			fail("synthetic parent path %q not found in resource %q schema", strings.Join(segs[:len(segs)-1], "."), resName)
		}
		if !appendChild(parent, &attr) {
			fail("synthetic parent %q in resource %q is not a nested attribute", strings.Join(segs[:len(segs)-1], "."), resName)
		}
	}
}

// injectDSSyntheticAttributes adds `items.`-prefixed synthetic attributes to the
// matching datasources: the path (without the prefix) is resolved inside the
// datasource `items` element the same way the resource pass resolves dotted
// paths. A service without a datasource, or without `items.`-prefixed specs, is
// a no-op; declaring one with no datasource to receive it is a fail (typo
// guard). Runs right after the resource pass, before renames, so later passes
// see a complete schema on both sides.
func injectDSSyntheticAttributes(s *spec.Specification, resName string, specs []syntheticAttrSpec) {
	dsSpecs := collectDSSyntheticSpecs(resName, specs)
	if len(dsSpecs) == 0 {
		return
	}
	matched := false
	for i := range s.DataSources {
		ds := &s.DataSources[i]
		if ds.Name != resName || ds.Schema == nil {
			continue
		}
		matched = true
		for _, sp := range dsSpecs {
			injectDSSyntheticAttribute(ds, resName, &sp)
		}
	}
	if !matched {
		fail("%d `items.`-prefixed synthetic attribute(s) declared for resource %q, which has no datasource", len(dsSpecs), resName)
	}
}

func collectDSSyntheticSpecs(resName string, specs []syntheticAttrSpec) []syntheticAttrSpec {
	var dsSpecs []syntheticAttrSpec
	for _, sp := range specs {
		if !strings.HasPrefix(sp.Path, "items.") {
			continue
		}
		if strings.Count(sp.Path, "items.") > 1 || strings.HasPrefix(strings.TrimPrefix(sp.Path, "items."), "items.") {
			fail("synthetic path %q in datasource %q: nested %q prefixes are not supported", sp.Path, resName, "items.")
		}
		sp.Path = strings.TrimPrefix(sp.Path, "items.")
		dsSpecs = append(dsSpecs, sp)
	}
	return dsSpecs
}

func injectDSSyntheticAttribute(ds *datasource.DataSource, resName string, sp *syntheticAttrSpec) {
	if sp.Path == "" {
		fail("synthetic attribute with empty path in datasource %q", resName)
	}
	segs := strings.Split(sp.Path, ".")
	leaf := segs[len(segs)-1]
	attr, err := buildSyntheticAttribute(leaf, sp)
	if err != nil {
		fail("synthetic attribute %q in datasource %q: %v", sp.Path, resName, err)
	}
	dsAttr := resourceAttrToDS(&attr)
	dsSegs := append([]string{datasourceItemsSegment}, segs...)
	if resolveDSAttribute(ds.Schema.Attributes, dsSegs) != nil {
		fail("synthetic path %q already exists in datasource %q schema", sp.Path, resName)
	}
	parent := resolveDSAttribute(ds.Schema.Attributes, dsSegs[:len(dsSegs)-1])
	if parent == nil {
		fail("synthetic parent path %q not found in datasource %q schema", strings.Join(dsSegs[:len(dsSegs)-1], "."), resName)
	}
	if !appendDSChild(parent, &dsAttr) {
		fail("synthetic parent %q in datasource %q is not a nested attribute", strings.Join(dsSegs[:len(dsSegs)-1], "."), resName)
	}
}

type syntheticBuilder func(a *resource.Attribute, name string, sp *syntheticAttrSpec, cor schema.ComputedOptionalRequired, desc, dep *string) error

var syntheticBuilders map[string]syntheticBuilder

func init() {
	syntheticBuilders = map[string]syntheticBuilder{
		"string":        buildSyntheticString,
		"bool":          buildSyntheticBool,
		"int64":         buildSyntheticInt64,
		"float64":       buildSyntheticFloat64,
		"number":        buildSyntheticNumber,
		"list":          buildSyntheticList,
		"single_nested": buildSyntheticSingleNested,
	}
}

// buildSyntheticAttribute constructs a resource.Attribute from its spec. Only
// the structural fields (type, optionality, description, deprecation, element
// type, nested children) are set here; plan modifiers / validators / defaults
// are applied by their own passes, which run afterwards.
func buildSyntheticAttribute(name string, sp *syntheticAttrSpec) (resource.Attribute, error) {
	cor, err := parseCOR(sp.Value)
	if err != nil {
		return resource.Attribute{}, err
	}
	desc := optStr(sp.Description)
	dep := optStr(sp.DeprecationMessage)
	a := resource.Attribute{Name: name}
	builder, ok := syntheticBuilders[sp.Type]
	if !ok {
		if sp.Type == "" {
			return resource.Attribute{}, errMissingType
		}
		return resource.Attribute{}, fmt.Errorf("%w %q", errUnsupportedSyntheticType, sp.Type)
	}
	if err := builder(&a, name, sp, cor, desc, dep); err != nil {
		return resource.Attribute{}, err
	}
	return a, nil
}

func buildSyntheticString(a *resource.Attribute, _ string, _ *syntheticAttrSpec, cor schema.ComputedOptionalRequired, desc, dep *string) error {
	a.String = &resource.StringAttribute{ComputedOptionalRequired: cor, Description: desc, DeprecationMessage: dep}
	return nil
}

func buildSyntheticBool(a *resource.Attribute, _ string, _ *syntheticAttrSpec, cor schema.ComputedOptionalRequired, desc, dep *string) error {
	a.Bool = &resource.BoolAttribute{ComputedOptionalRequired: cor, Description: desc, DeprecationMessage: dep}
	return nil
}

func buildSyntheticInt64(a *resource.Attribute, _ string, _ *syntheticAttrSpec, cor schema.ComputedOptionalRequired, desc, dep *string) error {
	a.Int64 = &resource.Int64Attribute{ComputedOptionalRequired: cor, Description: desc, DeprecationMessage: dep}
	return nil
}

func buildSyntheticFloat64(a *resource.Attribute, _ string, _ *syntheticAttrSpec, cor schema.ComputedOptionalRequired, desc, dep *string) error {
	a.Float64 = &resource.Float64Attribute{ComputedOptionalRequired: cor, Description: desc, DeprecationMessage: dep}
	return nil
}

func buildSyntheticNumber(a *resource.Attribute, _ string, _ *syntheticAttrSpec, cor schema.ComputedOptionalRequired, desc, dep *string) error {
	a.Number = &resource.NumberAttribute{ComputedOptionalRequired: cor, Description: desc, DeprecationMessage: dep}
	return nil
}

func buildSyntheticList(a *resource.Attribute, _ string, sp *syntheticAttrSpec, cor schema.ComputedOptionalRequired, desc, dep *string) error {
	et, err := parseElementType(sp.ElementType)
	if err != nil {
		return err
	}
	a.List = &resource.ListAttribute{ComputedOptionalRequired: cor, ElementType: et, Description: desc, DeprecationMessage: dep}
	return nil
}

func buildSyntheticSingleNested(
	a *resource.Attribute,
	name string,
	sp *syntheticAttrSpec,
	cor schema.ComputedOptionalRequired,
	desc, dep *string,
) error {
	children, err := buildSyntheticChildren(name, sp.Attributes)
	if err != nil {
		return err
	}
	a.SingleNested = &resource.SingleNestedAttribute{
		ComputedOptionalRequired: cor,
		Attributes:               children,
		Description:              desc,
		DeprecationMessage:       dep,
	}
	return nil
}

func buildSyntheticChildren(name string, specs []syntheticAttrSpec) (resource.Attributes, error) {
	children := make(resource.Attributes, 0, len(specs))
	for _, c := range specs {
		if c.Path == "" {
			return nil, fmt.Errorf("%w under %q", errNestedAttributeEmptyPath, name)
		}
		seg := strings.Split(c.Path, ".")
		child, err := buildSyntheticAttribute(seg[len(seg)-1], &c)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}

// resourceAttrToDS converts a resource.Attribute to a datasource.Attribute —
// the inverse of dsAttrToResource, restricted to the types the synthetic
// builder can produce (scalars, list, single_nested). Validators are never
// carried over: datasource attributes are read-only, and the validators pass
// only targets the resource schema.
func resourceAttrToDS(a *resource.Attribute) datasource.Attribute {
	out := datasource.Attribute{Name: a.Name}
	switch {
	case a.Bool != nil:
		out.Bool = &datasource.BoolAttribute{
			ComputedOptionalRequired: a.Bool.ComputedOptionalRequired,
			CustomType:               a.Bool.CustomType,
			DeprecationMessage:       a.Bool.DeprecationMessage,
			Description:              a.Bool.Description,
			Sensitive:                a.Bool.Sensitive,
		}
	case a.Float64 != nil:
		out.Float64 = &datasource.Float64Attribute{
			ComputedOptionalRequired: a.Float64.ComputedOptionalRequired,
			CustomType:               a.Float64.CustomType,
			DeprecationMessage:       a.Float64.DeprecationMessage,
			Description:              a.Float64.Description,
			Sensitive:                a.Float64.Sensitive,
		}
	case a.Int64 != nil:
		out.Int64 = &datasource.Int64Attribute{
			ComputedOptionalRequired: a.Int64.ComputedOptionalRequired,
			CustomType:               a.Int64.CustomType,
			DeprecationMessage:       a.Int64.DeprecationMessage,
			Description:              a.Int64.Description,
			Sensitive:                a.Int64.Sensitive,
		}
	case a.Number != nil:
		out.Number = &datasource.NumberAttribute{
			ComputedOptionalRequired: a.Number.ComputedOptionalRequired,
			CustomType:               a.Number.CustomType,
			DeprecationMessage:       a.Number.DeprecationMessage,
			Description:              a.Number.Description,
			Sensitive:                a.Number.Sensitive,
		}
	case a.String != nil:
		out.String = &datasource.StringAttribute{
			ComputedOptionalRequired: a.String.ComputedOptionalRequired,
			CustomType:               a.String.CustomType,
			DeprecationMessage:       a.String.DeprecationMessage,
			Description:              a.String.Description,
			Sensitive:                a.String.Sensitive,
		}
	case a.List != nil:
		out.List = &datasource.ListAttribute{
			ComputedOptionalRequired: a.List.ComputedOptionalRequired,
			ElementType:              a.List.ElementType,
			CustomType:               a.List.CustomType,
			DeprecationMessage:       a.List.DeprecationMessage,
			Description:              a.List.Description,
			Sensitive:                a.List.Sensitive,
		}
	case a.SingleNested != nil:
		children := make(datasource.Attributes, 0, len(a.SingleNested.Attributes))
		for i := range a.SingleNested.Attributes {
			children = append(children, resourceAttrToDS(&a.SingleNested.Attributes[i]))
		}
		out.SingleNested = &datasource.SingleNestedAttribute{
			ComputedOptionalRequired: a.SingleNested.ComputedOptionalRequired,
			Attributes:               children,
			CustomType:               a.SingleNested.CustomType,
			DeprecationMessage:       a.SingleNested.DeprecationMessage,
			Description:              a.SingleNested.Description,
			Sensitive:                a.SingleNested.Sensitive,
		}
	default:
		fail("synthetic datasource projection: attribute %q has an unsupported type", a.Name)
	}
	return out
}

// appendDSChild appends child to the nested parent attribute (datasource side).
func appendDSChild(parent, child *datasource.Attribute) bool {
	switch {
	case parent.SingleNested != nil:
		parent.SingleNested.Attributes = append(parent.SingleNested.Attributes, *child)
	case parent.ListNested != nil:
		parent.ListNested.NestedObject.Attributes = append(parent.ListNested.NestedObject.Attributes, *child)
	case parent.SetNested != nil:
		parent.SetNested.NestedObject.Attributes = append(parent.SetNested.NestedObject.Attributes, *child)
	case parent.MapNested != nil:
		parent.MapNested.NestedObject.Attributes = append(parent.MapNested.NestedObject.Attributes, *child)
	default:
		return false
	}
	return true
}

func parseElementType(t string) (schema.ElementType, error) {
	switch t {
	case "string":
		return schema.ElementType{String: &schema.StringType{}}, nil
	case "bool":
		return schema.ElementType{Bool: &schema.BoolType{}}, nil
	case "int64":
		return schema.ElementType{Int64: &schema.Int64Type{}}, nil
	case "float64":
		return schema.ElementType{Float64: &schema.Float64Type{}}, nil
	case "number":
		return schema.ElementType{Number: &schema.NumberType{}}, nil
	case "":
		return schema.ElementType{}, errSyntheticListMissingElementType
	default:
		return schema.ElementType{}, fmt.Errorf("%w %q", errUnsupportedElementType, t)
	}
}

// appendChild appends a child attribute to a nested parent's attribute set.
func appendChild(parent, child *resource.Attribute) bool {
	switch {
	case parent.SingleNested != nil:
		parent.SingleNested.Attributes = append(parent.SingleNested.Attributes, *child)
	case parent.ListNested != nil:
		parent.ListNested.NestedObject.Attributes = append(parent.ListNested.NestedObject.Attributes, *child)
	case parent.SetNested != nil:
		parent.SetNested.NestedObject.Attributes = append(parent.SetNested.NestedObject.Attributes, *child)
	case parent.MapNested != nil:
		parent.MapNested.NestedObject.Attributes = append(parent.MapNested.NestedObject.Attributes, *child)
	default:
		return false
	}
	return true
}

func optStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ── renames (collision-avoiding leaf renames) ───────────────────────

// injectRenames renames each declared leaf in the resource schema (at Path) and
// in the matching datasource schema (at "items."+Path). At least one side must
// match (typo guard). Renaming only changes the attribute name, so the framework
// generator emits a distinct Go type and no longer collides with another
// occurrence of the OAS name in a different nesting context.
func injectRenames(s *spec.Specification, res *resource.Resource, resName string, specs []renameSpec) {
	for _, r := range specs {
		if r.Path == "" || r.To == "" {
			fail("rename in resource %q needs both path and to", resName)
		}
		segs := strings.Split(r.Path, ".")
		applied := false
		if a := resolveAttribute(res.Schema.Attributes, segs); a != nil {
			a.Name = r.To
			applied = true
		}
		for i := range s.DataSources {
			if s.DataSources[i].Name != resName || s.DataSources[i].Schema == nil {
				continue
			}
			dsSegs := append([]string{datasourceItemsSegment}, segs...)
			if a := resolveDSAttribute(s.DataSources[i].Schema.Attributes, dsSegs); a != nil {
				a.Name = r.To
				applied = true
			}
		}
		if !applied {
			fail("rename path %q not found in resource or datasource %q", r.Path, resName)
		}
	}
}

// resolveDSAttribute is the datasource counterpart of resolveAttribute.
func resolveDSAttribute(attrs datasource.Attributes, segments []string) *datasource.Attribute {
	if len(segments) == 0 {
		return nil
	}
	for i := range attrs {
		if attrs[i].Name != segments[0] {
			continue
		}
		if len(segments) == 1 {
			return &attrs[i]
		}
		child := dsChildAttributes(&attrs[i])
		if child == nil {
			return nil
		}
		return resolveDSAttribute(child, segments[1:])
	}
	return nil
}

func dsChildAttributes(a *datasource.Attribute) datasource.Attributes {
	switch {
	case a.ListNested != nil:
		return a.ListNested.NestedObject.Attributes
	case a.SetNested != nil:
		return a.SetNested.NestedObject.Attributes
	case a.MapNested != nil:
		return a.MapNested.NestedObject.Attributes
	case a.SingleNested != nil:
		return a.SingleNested.Attributes
	}
	return nil
}

// ── create-without-body resource synthesis ──────────────────────────

// synthesizeResourceFromItems builds a resource schema for a create-without-body
// resource (core.create.no_request_body) by converting the datasource `items`
// nested object — the per-instance read view — into a resource schema. The
// converted attributes are all `computed` (as the read schema is); the
// optionality / plan-modifier / validator passes that run afterwards mark the
// user-settable ones (e.g. vpc_id, vm_id, nic_id) and add their behavior. The
// synthesized resource is appended to the spec so `make generate-spec` produces
// the same IR every time (the first regen becomes the drift baseline).
func synthesizeResourceFromItems(s *spec.Specification, name string) (*resource.Resource, error) {
	ds, err := findDataSourceByName(s, name)
	if err != nil {
		return nil, err
	}
	items, err := findItemsAttribute(ds, name)
	if err != nil {
		return nil, err
	}
	nested, err := itemsNestedAttributes(items, name)
	if err != nil {
		return nil, err
	}
	attrs := make(resource.Attributes, 0, len(nested))
	for i := range nested {
		attrs = append(attrs, dsAttrToResource(&nested[i]))
	}
	s.Resources = append(s.Resources, resource.Resource{
		Name:   name,
		Schema: &resource.Schema{Attributes: attrs},
	})
	return &s.Resources[len(s.Resources)-1], nil
}

func findDataSourceByName(s *spec.Specification, name string) (*datasource.DataSource, error) {
	for i := range s.DataSources {
		if s.DataSources[i].Name == name && s.DataSources[i].Schema != nil {
			return &s.DataSources[i], nil
		}
	}
	return nil, fmt.Errorf("%w %q (with schema)", errDatasourceNotFound, name)
}

func findItemsAttribute(ds *datasource.DataSource, name string) (*datasource.Attribute, error) {
	for i := range ds.Schema.Attributes {
		if ds.Schema.Attributes[i].Name == datasourceItemsSegment {
			return &ds.Schema.Attributes[i], nil
		}
	}
	return nil, fmt.Errorf("%w %q", errDatasourceItemsMissing, name)
}

func itemsNestedAttributes(items *datasource.Attribute, name string) (datasource.Attributes, error) {
	switch {
	case items.ListNested != nil:
		return items.ListNested.NestedObject.Attributes, nil
	case items.SetNested != nil:
		return items.SetNested.NestedObject.Attributes, nil
	default:
		return nil, fmt.Errorf("%w %q \"items\"", errDatasourceItemsNotCollection, name)
	}
}

// dsAttrToResource converts a datasource.Attribute to a resource.Attribute. The
// structural fields (type, optionality, element type, description, deprecation,
// sensitivity, custom type, nested children) carry over; datasource-only
// validators are dropped (the resource's are declared via the validators: pass).
func dsAttrToResource(a *datasource.Attribute) resource.Attribute {
	out := resource.Attribute{Name: a.Name}
	converters := []func(*datasource.Attribute, *resource.Attribute) bool{
		convertDSBoolAttr,
		convertDSStringAttr,
		convertDSInt64Attr,
		convertDSFloat64Attr,
		convertDSNumberAttr,
		convertDSListAttr,
		convertDSSetAttr,
		convertDSMapAttr,
		convertDSObjectAttr,
		convertDSListNestedAttr,
		convertDSSetNestedAttr,
		convertDSMapNestedAttr,
		convertDSSingleNestedAttr,
	}
	for _, convert := range converters {
		if convert(a, &out) {
			return out
		}
	}
	fail("create-without-body synthesis: attribute %q has an unsupported type", a.Name)
	return out
}

func convertDSBoolAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.Bool == nil {
		return false
	}
	out.Bool = &resource.BoolAttribute{
		ComputedOptionalRequired: a.Bool.ComputedOptionalRequired,
		CustomType:               a.Bool.CustomType,
		DeprecationMessage:       a.Bool.DeprecationMessage,
		Description:              a.Bool.Description,
		Sensitive:                a.Bool.Sensitive,
	}
	return true
}

func convertDSStringAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.String == nil {
		return false
	}
	out.String = &resource.StringAttribute{
		ComputedOptionalRequired: a.String.ComputedOptionalRequired,
		CustomType:               a.String.CustomType,
		DeprecationMessage:       a.String.DeprecationMessage,
		Description:              a.String.Description,
		Sensitive:                a.String.Sensitive,
	}
	return true
}

func convertDSInt64Attr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.Int64 == nil {
		return false
	}
	out.Int64 = &resource.Int64Attribute{
		ComputedOptionalRequired: a.Int64.ComputedOptionalRequired,
		CustomType:               a.Int64.CustomType,
		DeprecationMessage:       a.Int64.DeprecationMessage,
		Description:              a.Int64.Description,
		Sensitive:                a.Int64.Sensitive,
	}
	return true
}

func convertDSFloat64Attr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.Float64 == nil {
		return false
	}
	out.Float64 = &resource.Float64Attribute{
		ComputedOptionalRequired: a.Float64.ComputedOptionalRequired,
		CustomType:               a.Float64.CustomType,
		DeprecationMessage:       a.Float64.DeprecationMessage,
		Description:              a.Float64.Description,
		Sensitive:                a.Float64.Sensitive,
	}
	return true
}

func convertDSNumberAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.Number == nil {
		return false
	}
	out.Number = &resource.NumberAttribute{
		ComputedOptionalRequired: a.Number.ComputedOptionalRequired,
		CustomType:               a.Number.CustomType,
		DeprecationMessage:       a.Number.DeprecationMessage,
		Description:              a.Number.Description,
		Sensitive:                a.Number.Sensitive,
	}
	return true
}

func convertDSListAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.List == nil {
		return false
	}
	out.List = &resource.ListAttribute{
		ComputedOptionalRequired: a.List.ComputedOptionalRequired,
		ElementType:              a.List.ElementType,
		CustomType:               a.List.CustomType,
		DeprecationMessage:       a.List.DeprecationMessage,
		Description:              a.List.Description,
		Sensitive:                a.List.Sensitive,
	}
	return true
}

func convertDSSetAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.Set == nil {
		return false
	}
	out.Set = &resource.SetAttribute{
		ComputedOptionalRequired: a.Set.ComputedOptionalRequired,
		ElementType:              a.Set.ElementType,
		CustomType:               a.Set.CustomType,
		DeprecationMessage:       a.Set.DeprecationMessage,
		Description:              a.Set.Description,
		Sensitive:                a.Set.Sensitive,
	}
	return true
}

func convertDSMapAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.Map == nil {
		return false
	}
	out.Map = &resource.MapAttribute{
		ComputedOptionalRequired: a.Map.ComputedOptionalRequired,
		ElementType:              a.Map.ElementType,
		CustomType:               a.Map.CustomType,
		DeprecationMessage:       a.Map.DeprecationMessage,
		Description:              a.Map.Description,
		Sensitive:                a.Map.Sensitive,
	}
	return true
}

func convertDSObjectAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.Object == nil {
		return false
	}
	out.Object = &resource.ObjectAttribute{
		AttributeTypes:           a.Object.AttributeTypes,
		ComputedOptionalRequired: a.Object.ComputedOptionalRequired,
		CustomType:               a.Object.CustomType,
		DeprecationMessage:       a.Object.DeprecationMessage,
		Description:              a.Object.Description,
		Sensitive:                a.Object.Sensitive,
	}
	return true
}

func convertDSListNestedAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.ListNested == nil {
		return false
	}
	out.ListNested = &resource.ListNestedAttribute{
		ComputedOptionalRequired: a.ListNested.ComputedOptionalRequired,
		NestedObject:             dsNestedToResource(a.ListNested.NestedObject),
		CustomType:               a.ListNested.CustomType,
		DeprecationMessage:       a.ListNested.DeprecationMessage,
		Description:              a.ListNested.Description,
		Sensitive:                a.ListNested.Sensitive,
	}
	return true
}

func convertDSSetNestedAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.SetNested == nil {
		return false
	}
	out.SetNested = &resource.SetNestedAttribute{
		ComputedOptionalRequired: a.SetNested.ComputedOptionalRequired,
		NestedObject:             dsNestedToResource(a.SetNested.NestedObject),
		CustomType:               a.SetNested.CustomType,
		DeprecationMessage:       a.SetNested.DeprecationMessage,
		Description:              a.SetNested.Description,
		Sensitive:                a.SetNested.Sensitive,
	}
	return true
}

func convertDSMapNestedAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.MapNested == nil {
		return false
	}
	out.MapNested = &resource.MapNestedAttribute{
		ComputedOptionalRequired: a.MapNested.ComputedOptionalRequired,
		NestedObject:             dsNestedToResource(a.MapNested.NestedObject),
		CustomType:               a.MapNested.CustomType,
		DeprecationMessage:       a.MapNested.DeprecationMessage,
		Description:              a.MapNested.Description,
		Sensitive:                a.MapNested.Sensitive,
	}
	return true
}

func convertDSSingleNestedAttr(a *datasource.Attribute, out *resource.Attribute) bool {
	if a.SingleNested == nil {
		return false
	}
	children := make(resource.Attributes, 0, len(a.SingleNested.Attributes))
	for i := range a.SingleNested.Attributes {
		children = append(children, dsAttrToResource(&a.SingleNested.Attributes[i]))
	}
	out.SingleNested = &resource.SingleNestedAttribute{
		ComputedOptionalRequired: a.SingleNested.ComputedOptionalRequired,
		Attributes:               children,
		AssociatedExternalType:   a.SingleNested.AssociatedExternalType,
		CustomType:               a.SingleNested.CustomType,
		DeprecationMessage:       a.SingleNested.DeprecationMessage,
		Description:              a.SingleNested.Description,
		Sensitive:                a.SingleNested.Sensitive,
	}
	return true
}

func dsNestedToResource(n datasource.NestedAttributeObject) resource.NestedAttributeObject {
	children := make(resource.Attributes, 0, len(n.Attributes))
	for i := range n.Attributes {
		children = append(children, dsAttrToResource(&n.Attributes[i]))
	}
	return resource.NestedAttributeObject{
		Attributes:             children,
		AssociatedExternalType: n.AssociatedExternalType,
		CustomType:             n.CustomType,
	}
}

// ── sets (list -> set conversion) ───────────────────────────────────

func injectSets(res *resource.Resource, resName string, paths []string) {
	for _, p := range paths {
		if p == "" {
			fail("set with empty path in resource %q", resName)
		}
		attr := resolveAttribute(res.Schema.Attributes, strings.Split(p, "."))
		if attr == nil {
			fail("set path %q not found in resource %q schema", p, resName)
		}
		if !convertToSet(attr) {
			fail("attribute %q in resource %q is not a list/list_nested (cannot convert to set)", p, resName)
		}
	}
}

// convertToSet rewrites a list/list_nested attribute as set/set_nested,
// carrying over every field. It is idempotent (already-set attributes pass).
func convertToSet(a *resource.Attribute) bool {
	switch {
	case a.Set != nil, a.SetNested != nil:
		return true
	case a.List != nil:
		l := a.List
		a.Set = &resource.SetAttribute{
			ComputedOptionalRequired: l.ComputedOptionalRequired,
			ElementType:              l.ElementType,
			CustomType:               l.CustomType,
			DeprecationMessage:       l.DeprecationMessage,
			Description:              l.Description,
			Sensitive:                l.Sensitive,
			Default:                  listToSetDefault(l.Default),
			PlanModifiers:            listToSetPlanModifiers(l.PlanModifiers),
			Validators:               listToSetValidators(l.Validators),
		}
		a.List = nil
		return true
	case a.ListNested != nil:
		l := a.ListNested
		a.SetNested = &resource.SetNestedAttribute{
			ComputedOptionalRequired: l.ComputedOptionalRequired,
			NestedObject:             l.NestedObject,
			CustomType:               l.CustomType,
			DeprecationMessage:       l.DeprecationMessage,
			Description:              l.Description,
			Sensitive:                l.Sensitive,
			Default:                  listToSetDefault(l.Default),
			PlanModifiers:            listToSetPlanModifiers(l.PlanModifiers),
			Validators:               listToSetValidators(l.Validators),
		}
		a.ListNested = nil
		return true
	}
	return false
}

func listToSetDefault(d *schema.ListDefault) *schema.SetDefault {
	if d == nil {
		return nil
	}
	return &schema.SetDefault{Custom: d.Custom}
}

func listToSetPlanModifiers(in schema.ListPlanModifiers) schema.SetPlanModifiers {
	if len(in) == 0 {
		return nil
	}
	out := make(schema.SetPlanModifiers, len(in))
	for i, m := range in {
		out[i] = schema.SetPlanModifier(m)
	}
	return out
}

func listToSetValidators(in schema.ListValidators) schema.SetValidators {
	if len(in) == 0 {
		return nil
	}
	out := make(schema.SetValidators, len(in))
	for i, v := range in {
		out[i] = schema.SetValidator(v)
	}
	return out
}

// ── optionality (computed_optional_required) ────────────────────────

func injectOptionality(res *resource.Resource, resName string, specs []optionalitySpec) {
	for _, o := range specs {
		if o.Path == "" {
			fail("optionality with empty path in resource %q", resName)
		}
		cor, err := parseCOR(o.Value)
		if err != nil {
			fail("optionality on %q in resource %q: %v", o.Path, resName, err)
		}
		attr := resolveAttribute(res.Schema.Attributes, strings.Split(o.Path, "."))
		if attr == nil {
			fail("optionality path %q not found in resource %q schema", o.Path, resName)
		}
		if !setComputedOptionalRequired(attr, cor) {
			fail("attribute %q in resource %q has no optionality-capable type", o.Path, resName)
		}
	}
}

func parseCOR(v string) (schema.ComputedOptionalRequired, error) {
	switch v {
	case "computed":
		return schema.Computed, nil
	case "computed_optional":
		return schema.ComputedOptional, nil
	case "optional":
		return schema.Optional, nil
	case "required":
		return schema.Required, nil
	default:
		return "", fmt.Errorf("%w %q (want computed|computed_optional|optional|required)", errInvalidOptionality, v)
	}
}

func setComputedOptionalRequired(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	setters := []func(*resource.Attribute, schema.ComputedOptionalRequired) bool{
		setResourceBoolCOR,
		setResourceStringCOR,
		setResourceInt64COR,
		setResourceFloat64COR,
		setResourceNumberCOR,
		setResourceListCOR,
		setResourceListNestedCOR,
		setResourceSetCOR,
		setResourceSetNestedCOR,
		setResourceMapCOR,
		setResourceMapNestedCOR,
		setResourceObjectCOR,
		setResourceSingleNestedCOR,
	}
	for _, set := range setters {
		if set(a, cor) {
			return true
		}
	}
	return false
}

func setResourceBoolCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.Bool == nil {
		return false
	}
	a.Bool.ComputedOptionalRequired = cor
	return true
}

func setResourceStringCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.String == nil {
		return false
	}
	a.String.ComputedOptionalRequired = cor
	return true
}

func setResourceInt64COR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.Int64 == nil {
		return false
	}
	a.Int64.ComputedOptionalRequired = cor
	return true
}

func setResourceFloat64COR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.Float64 == nil {
		return false
	}
	a.Float64.ComputedOptionalRequired = cor
	return true
}

func setResourceNumberCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.Number == nil {
		return false
	}
	a.Number.ComputedOptionalRequired = cor
	return true
}

func setResourceListCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.List == nil {
		return false
	}
	a.List.ComputedOptionalRequired = cor
	return true
}

func setResourceListNestedCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.ListNested == nil {
		return false
	}
	a.ListNested.ComputedOptionalRequired = cor
	return true
}

func setResourceSetCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.Set == nil {
		return false
	}
	a.Set.ComputedOptionalRequired = cor
	return true
}

func setResourceSetNestedCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.SetNested == nil {
		return false
	}
	a.SetNested.ComputedOptionalRequired = cor
	return true
}

func setResourceMapCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.Map == nil {
		return false
	}
	a.Map.ComputedOptionalRequired = cor
	return true
}

func setResourceMapNestedCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.MapNested == nil {
		return false
	}
	a.MapNested.ComputedOptionalRequired = cor
	return true
}

func setResourceObjectCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.Object == nil {
		return false
	}
	a.Object.ComputedOptionalRequired = cor
	return true
}

func setResourceSingleNestedCOR(a *resource.Attribute, cor schema.ComputedOptionalRequired) bool {
	if a.SingleNested == nil {
		return false
	}
	a.SingleNested.ComputedOptionalRequired = cor
	return true
}

// ── plan modifiers ──────────────────────────────────────────────────

// mergedPlanModifiers expands the requires_replace shorthand and appends the
// explicit plan_modifiers list, preserving declaration order (shorthand first).
func mergedPlanModifiers(d *descriptorHead) []planModifierSpec {
	out := make([]planModifierSpec, 0, len(d.RequiresReplace)+len(d.PlanModifiers))
	for _, name := range d.RequiresReplace {
		out = append(out, planModifierSpec{Path: name, Kind: kindRequiresReplace})
	}
	out = append(out, d.PlanModifiers...)
	return out
}

func injectPlanModifiers(res *resource.Resource, resName string, specs []planModifierSpec) {
	type group struct {
		segments []string
		specs    []planModifierSpec
	}
	groups := map[string]*group{}
	var order []string
	for _, m := range specs {
		if m.Path == "" {
			fail("plan modifier with empty path in resource %q", resName)
		}
		g, ok := groups[m.Path]
		if !ok {
			g = &group{segments: strings.Split(m.Path, ".")}
			groups[m.Path] = g
			order = append(order, m.Path)
		}
		g.specs = append(g.specs, m)
	}
	for _, p := range order {
		g := groups[p]
		attr := resolveAttribute(res.Schema.Attributes, g.segments)
		if attr == nil {
			fail("plan modifier path %q not found in resource %q schema", p, resName)
		}
		sub, ok := subpackageFor(attr)
		if !ok {
			fail("attribute %q in resource %q has no plan-modifier-capable type", p, resName)
		}
		customs := make([]schema.CustomPlanModifier, 0, len(g.specs))
		for _, m := range g.specs {
			def, err := planModifierDefinition(sub, &m)
			if err != nil {
				fail("path %q: %v", p, err)
			}
			customs = append(customs, customPlanModifier(sub, def))
		}
		setPlanModifiers(attr, customs)
	}
}

// subpackageFor returns the planmodifier subpackage matching the attribute's type.
func subpackageFor(a *resource.Attribute) (string, bool) {
	switch {
	case a.Bool != nil:
		return "boolplanmodifier", true
	case a.String != nil:
		return "stringplanmodifier", true
	case a.Int64 != nil:
		return "int64planmodifier", true
	case a.Float64 != nil:
		return "float64planmodifier", true
	case a.Number != nil:
		return "numberplanmodifier", true
	case a.List != nil, a.ListNested != nil:
		return "listplanmodifier", true
	case a.Set != nil, a.SetNested != nil:
		return "setplanmodifier", true
	case a.Map != nil, a.MapNested != nil:
		return "mapplanmodifier", true
	case a.Object != nil, a.SingleNested != nil:
		return "objectplanmodifier", true
	}
	return "", false
}

func planModifierDefinition(sub string, m *planModifierSpec) (string, error) {
	switch m.Kind {
	case kindRequiresReplace:
		return sub + ".RequiresReplace()", nil
	case kindRequiresReplaceIfConfigured:
		return sub + ".RequiresReplaceIfConfigured()", nil
	case kindUseStateForUnknown:
		return sub + ".UseStateForUnknown()", nil
	case kindRequiresReplaceIf:
		if m.Func == "" || m.Description == "" {
			return "", fmt.Errorf("%w %q", errRequiresReplaceIfMissingFields, kindRequiresReplaceIf)
		}
		md := m.DescriptionMarkdown
		if md == "" {
			md = m.Description
		}
		return fmt.Sprintf("%s.RequiresReplaceIf(%s, %s, %s)", sub, m.Func, strconv.Quote(m.Description), strconv.Quote(md)), nil
	case "":
		return "", errMissingPlanModifierKind
	default:
		return "", fmt.Errorf("%w %q", errUnknownPlanModifierKind, m.Kind)
	}
}

func customPlanModifier(sub, schemaDef string) schema.CustomPlanModifier {
	return schema.CustomPlanModifier{
		Imports: []code.Import{
			{Path: planModifierBaseImport},
			{Path: planModifierSubImportPrefix + sub},
		},
		SchemaDefinition: schemaDef,
	}
}

// setPlanModifiers replaces the attribute's plan_modifiers with the given custom
// modifiers. Validators, defaults and every other field are left untouched.
func setPlanModifiers(a *resource.Attribute, customs []schema.CustomPlanModifier) {
	ptrs := make([]*schema.CustomPlanModifier, len(customs))
	for i := range customs {
		ptrs[i] = &customs[i]
	}
	setters := []func(*resource.Attribute, []*schema.CustomPlanModifier) bool{
		setResourceBoolPlanModifiers,
		setResourceStringPlanModifiers,
		setResourceInt64PlanModifiers,
		setResourceFloat64PlanModifiers,
		setResourceNumberPlanModifiers,
		setResourceListPlanModifiers,
		setResourceListNestedPlanModifiers,
		setResourceSetPlanModifiers,
		setResourceSetNestedPlanModifiers,
		setResourceMapPlanModifiers,
		setResourceMapNestedPlanModifiers,
		setResourceObjectPlanModifiers,
		setResourceSingleNestedPlanModifiers,
	}
	for _, set := range setters {
		if set(a, ptrs) {
			return
		}
	}
}

func setResourceBoolPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.Bool == nil {
		return false
	}
	pm := make(schema.BoolPlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.BoolPlanModifier{Custom: c}
	}
	a.Bool.PlanModifiers = pm
	return true
}

func setResourceStringPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.String == nil {
		return false
	}
	pm := make(schema.StringPlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.StringPlanModifier{Custom: c}
	}
	a.String.PlanModifiers = pm
	return true
}

func setResourceInt64PlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.Int64 == nil {
		return false
	}
	pm := make(schema.Int64PlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.Int64PlanModifier{Custom: c}
	}
	a.Int64.PlanModifiers = pm
	return true
}

func setResourceFloat64PlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.Float64 == nil {
		return false
	}
	pm := make(schema.Float64PlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.Float64PlanModifier{Custom: c}
	}
	a.Float64.PlanModifiers = pm
	return true
}

func setResourceNumberPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.Number == nil {
		return false
	}
	pm := make(schema.NumberPlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.NumberPlanModifier{Custom: c}
	}
	a.Number.PlanModifiers = pm
	return true
}

func setResourceListPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.List == nil {
		return false
	}
	a.List.PlanModifiers = listPlanModifiers(ptrs)
	return true
}

func setResourceListNestedPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.ListNested == nil {
		return false
	}
	a.ListNested.PlanModifiers = listPlanModifiers(ptrs)
	return true
}

func setResourceSetPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.Set == nil {
		return false
	}
	a.Set.PlanModifiers = setPlanModifiersSlice(ptrs)
	return true
}

func setResourceSetNestedPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.SetNested == nil {
		return false
	}
	a.SetNested.PlanModifiers = setPlanModifiersSlice(ptrs)
	return true
}

func setResourceMapPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.Map == nil {
		return false
	}
	a.Map.PlanModifiers = mapPlanModifiers(ptrs)
	return true
}

func setResourceMapNestedPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.MapNested == nil {
		return false
	}
	a.MapNested.PlanModifiers = mapPlanModifiers(ptrs)
	return true
}

func setResourceObjectPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.Object == nil {
		return false
	}
	a.Object.PlanModifiers = objectPlanModifiers(ptrs)
	return true
}

func setResourceSingleNestedPlanModifiers(a *resource.Attribute, ptrs []*schema.CustomPlanModifier) bool {
	if a.SingleNested == nil {
		return false
	}
	a.SingleNested.PlanModifiers = objectPlanModifiers(ptrs)
	return true
}

func listPlanModifiers(ptrs []*schema.CustomPlanModifier) schema.ListPlanModifiers {
	pm := make(schema.ListPlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.ListPlanModifier{Custom: c}
	}
	return pm
}

func setPlanModifiersSlice(ptrs []*schema.CustomPlanModifier) schema.SetPlanModifiers {
	pm := make(schema.SetPlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.SetPlanModifier{Custom: c}
	}
	return pm
}

func mapPlanModifiers(ptrs []*schema.CustomPlanModifier) schema.MapPlanModifiers {
	pm := make(schema.MapPlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.MapPlanModifier{Custom: c}
	}
	return pm
}

func objectPlanModifiers(ptrs []*schema.CustomPlanModifier) schema.ObjectPlanModifiers {
	pm := make(schema.ObjectPlanModifiers, len(ptrs))
	for i, c := range ptrs {
		pm[i] = schema.ObjectPlanModifier{Custom: c}
	}
	return pm
}

// ── validators ──────────────────────────────────────────────────────

func injectValidators(res *resource.Resource, resName string, specs []validatorSpec) {
	for _, v := range specs {
		if v.Path == "" {
			fail("validator with empty path in resource %q", resName)
		}
		if v.SchemaDefinition == "" {
			fail("validator on %q in resource %q has no schema_definition", v.Path, resName)
		}
		attr := resolveAttribute(res.Schema.Attributes, strings.Split(v.Path, "."))
		if attr == nil {
			fail("validator path %q not found in resource %q schema", v.Path, resName)
		}
		imports := make([]code.Import, 0, len(v.Imports))
		for _, p := range v.Imports {
			imports = append(imports, code.Import{Path: p})
		}
		cv := &schema.CustomValidator{Imports: imports, SchemaDefinition: v.SchemaDefinition}
		if !appendValidator(attr, cv) {
			fail("attribute %q in resource %q has no validator-capable type", v.Path, resName)
		}
	}
}

// appendValidator appends the custom validator unless one with the same
// schema_definition is already present (idempotent), preserving OAS-derived ones.
func appendValidator(a *resource.Attribute, cv *schema.CustomValidator) bool {
	def := cv.SchemaDefinition
	appenders := []func(*resource.Attribute, *schema.CustomValidator, string) bool{
		appendResourceBoolValidator,
		appendResourceStringValidator,
		appendResourceInt64Validator,
		appendResourceFloat64Validator,
		appendResourceNumberValidator,
		appendResourceListValidator,
		appendResourceListNestedValidator,
		appendResourceSetValidator,
		appendResourceSetNestedValidator,
		appendResourceMapValidator,
		appendResourceMapNestedValidator,
		appendResourceObjectValidator,
		appendResourceSingleNestedValidator,
	}
	for _, app := range appenders {
		if app(a, cv, def) {
			return true
		}
	}
	return false
}

func appendResourceBoolValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.Bool == nil {
		return false
	}
	for _, e := range a.Bool.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.Bool.Validators = append(a.Bool.Validators, schema.BoolValidator{Custom: cv})
	return true
}

func appendResourceStringValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.String == nil {
		return false
	}
	for _, e := range a.String.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.String.Validators = append(a.String.Validators, schema.StringValidator{Custom: cv})
	return true
}

func appendResourceInt64Validator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.Int64 == nil {
		return false
	}
	for _, e := range a.Int64.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.Int64.Validators = append(a.Int64.Validators, schema.Int64Validator{Custom: cv})
	return true
}

func appendResourceFloat64Validator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.Float64 == nil {
		return false
	}
	for _, e := range a.Float64.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.Float64.Validators = append(a.Float64.Validators, schema.Float64Validator{Custom: cv})
	return true
}

func appendResourceNumberValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.Number == nil {
		return false
	}
	for _, e := range a.Number.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.Number.Validators = append(a.Number.Validators, schema.NumberValidator{Custom: cv})
	return true
}

func appendResourceListValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.List == nil {
		return false
	}
	for _, e := range a.List.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.List.Validators = append(a.List.Validators, schema.ListValidator{Custom: cv})
	return true
}

func appendResourceListNestedValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.ListNested == nil {
		return false
	}
	for _, e := range a.ListNested.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.ListNested.Validators = append(a.ListNested.Validators, schema.ListValidator{Custom: cv})
	return true
}

func appendResourceSetValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.Set == nil {
		return false
	}
	for _, e := range a.Set.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.Set.Validators = append(a.Set.Validators, schema.SetValidator{Custom: cv})
	return true
}

func appendResourceSetNestedValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.SetNested == nil {
		return false
	}
	for _, e := range a.SetNested.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.SetNested.Validators = append(a.SetNested.Validators, schema.SetValidator{Custom: cv})
	return true
}

func appendResourceMapValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.Map == nil {
		return false
	}
	for _, e := range a.Map.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.Map.Validators = append(a.Map.Validators, schema.MapValidator{Custom: cv})
	return true
}

func appendResourceMapNestedValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.MapNested == nil {
		return false
	}
	for _, e := range a.MapNested.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.MapNested.Validators = append(a.MapNested.Validators, schema.MapValidator{Custom: cv})
	return true
}

func appendResourceObjectValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.Object == nil {
		return false
	}
	for _, e := range a.Object.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.Object.Validators = append(a.Object.Validators, schema.ObjectValidator{Custom: cv})
	return true
}

func appendResourceSingleNestedValidator(a *resource.Attribute, cv *schema.CustomValidator, def string) bool {
	if a.SingleNested == nil {
		return false
	}
	for _, e := range a.SingleNested.Validators {
		if dup(e.Custom, def) {
			return true
		}
	}
	a.SingleNested.Validators = append(a.SingleNested.Validators, schema.ObjectValidator{Custom: cv})
	return true
}

func dup(c *schema.CustomValidator, def string) bool {
	return c != nil && c.SchemaDefinition == def
}

// ── defaults ────────────────────────────────────────────────────────

func injectDefaults(res *resource.Resource, resName string, specs []defaultSpec) {
	for _, d := range specs {
		if d.Path == "" {
			fail("default with empty path in resource %q", resName)
		}
		attr := resolveAttribute(res.Schema.Attributes, strings.Split(d.Path, "."))
		if attr == nil {
			fail("default path %q not found in resource %q schema", d.Path, resName)
		}
		if err := setDefault(attr, d.Static); err != nil {
			fail("default on %q in resource %q: %v", d.Path, resName, err)
		}
	}
}

func setDefault(a *resource.Attribute, static any) error {
	switch {
	case a.Bool != nil:
		b, ok := static.(bool)
		if !ok {
			return fmt.Errorf("%w, got %T", errBoolStaticMismatch, static)
		}
		a.Bool.Default = &schema.BoolDefault{Static: &b}
	case a.String != nil:
		s, ok := static.(string)
		if !ok {
			return fmt.Errorf("%w, got %T", errStringStaticMismatch, static)
		}
		a.String.Default = &schema.StringDefault{Static: &s}
	case a.Int64 != nil:
		i, err := toInt64(static)
		if err != nil {
			return err
		}
		a.Int64.Default = &schema.Int64Default{Static: &i}
	case a.Float64 != nil:
		f, err := toFloat64(static)
		if err != nil {
			return err
		}
		a.Float64.Default = &schema.Float64Default{Static: &f}
	default:
		return errStaticDefaultUnsupported
	}
	return nil
}

// clearDefaults removes any default from the named attributes. It is the inverse
// of injectDefaults: tfplugingen-openapi emits OAS-declared defaults the provider
// does not want (e.g. on a read-only mirror field), and there is no other way to
// drop them. Runs after injectDefaults so a path is never both set and cleared.
func clearDefaults(res *resource.Resource, resName string, paths []string) {
	for _, p := range paths {
		if p == "" {
			fail("clear_defaults with empty path in resource %q", resName)
		}
		attr := resolveAttribute(res.Schema.Attributes, strings.Split(p, "."))
		if attr == nil {
			fail("clear_defaults path %q not found in resource %q schema", p, resName)
		}
		if !clearDefault(attr) {
			fail("attribute %q in resource %q has no default-capable type", p, resName)
		}
	}
}

func clearDefault(a *resource.Attribute) bool {
	switch {
	case a.Bool != nil:
		a.Bool.Default = nil
	case a.String != nil:
		a.String.Default = nil
	case a.Int64 != nil:
		a.Int64.Default = nil
	case a.Float64 != nil:
		a.Float64.Default = nil
	default:
		return false
	}
	return true
}

// ── clear validators ────────────────────────────────────────────────

// applyToDSItems applies op to the datasource attribute at "items."+rest for
// each datasource matching resName. It mirrors the resource-side validation and
// failure messages for items.-prefixed paths.
func applyToDSItems(s *spec.Specification, resName, path string, op func(*datasource.Attribute) bool, opName, attrKind string) {
	rest := strings.TrimPrefix(path, "items.")
	applied := false
	for i := range s.DataSources {
		ds := &s.DataSources[i]
		if ds.Name != resName || ds.Schema == nil {
			continue
		}
		dsSegs := append([]string{datasourceItemsSegment}, strings.Split(rest, ".")...)
		if a := resolveDSAttribute(ds.Schema.Attributes, dsSegs); a != nil {
			if !op(a) {
				fail("%s path %q in datasource %q has no %s-capable type", opName, path, resName, attrKind)
			}
			applied = true
		}
	}
	if !applied {
		fail("%s path %q not found in datasource %q schema", opName, path, resName)
	}
}

// applyToResourceAttr applies op to the resource attribute at path, failing with
// the standard resource-side messages.
func applyToResourceAttr(res *resource.Resource, resName, path string, op func(*resource.Attribute) bool, opName, attrKind string) {
	attr := resolveAttribute(res.Schema.Attributes, strings.Split(path, "."))
	if attr == nil {
		fail("%s path %q not found in resource %q schema", opName, path, resName)
	}
	if !op(attr) {
		fail("attribute %q in resource %q has no %s-capable type", path, resName, attrKind)
	}
}

// clearValidators removes all validators from the named attributes. It is the
// inverse of injectValidators: tfplugingen-openapi derives validators from OAS
// enums (stringvalidator.OneOf), which the provider sometimes deliberately drops
// (e.g. a hardcoded AZ enum that would reject any future Subregion). A
// `items.`-prefixed path targets the matching datasource leaf (mirroring
// renames). Runs after injectValidators so a path is never both appended and
// cleared.
func clearValidators(s *spec.Specification, res *resource.Resource, resName string, paths []string) {
	for _, p := range paths {
		if p == "" {
			fail("clear_validators with empty path in resource %q", resName)
		}
		if strings.HasPrefix(p, "items.") {
			applyToDSItems(s, resName, p, clearDSValidators, "clear_validators", "validator")
			continue
		}
		applyToResourceAttr(res, resName, p, clearResourceValidators, "clear_validators", "validator")
	}
}

func clearResourceValidators(a *resource.Attribute) bool {
	clearers := []func(*resource.Attribute) bool{
		clearResourceBoolValidators,
		clearResourceStringValidators,
		clearResourceInt64Validators,
		clearResourceFloat64Validators,
		clearResourceNumberValidators,
		clearResourceListValidators,
		clearResourceListNestedValidators,
		clearResourceSetValidators,
		clearResourceSetNestedValidators,
		clearResourceMapValidators,
		clearResourceMapNestedValidators,
		clearResourceObjectValidators,
		clearResourceSingleNestedValidators,
	}
	for _, clear := range clearers {
		if clear(a) {
			return true
		}
	}
	return false
}

func clearResourceBoolValidators(a *resource.Attribute) bool {
	if a.Bool == nil {
		return false
	}
	a.Bool.Validators = nil
	return true
}

func clearResourceStringValidators(a *resource.Attribute) bool {
	if a.String == nil {
		return false
	}
	a.String.Validators = nil
	return true
}

func clearResourceInt64Validators(a *resource.Attribute) bool {
	if a.Int64 == nil {
		return false
	}
	a.Int64.Validators = nil
	return true
}

func clearResourceFloat64Validators(a *resource.Attribute) bool {
	if a.Float64 == nil {
		return false
	}
	a.Float64.Validators = nil
	return true
}

func clearResourceNumberValidators(a *resource.Attribute) bool {
	if a.Number == nil {
		return false
	}
	a.Number.Validators = nil
	return true
}

func clearResourceListValidators(a *resource.Attribute) bool {
	if a.List == nil {
		return false
	}
	a.List.Validators = nil
	return true
}

func clearResourceListNestedValidators(a *resource.Attribute) bool {
	if a.ListNested == nil {
		return false
	}
	a.ListNested.Validators = nil
	return true
}

func clearResourceSetValidators(a *resource.Attribute) bool {
	if a.Set == nil {
		return false
	}
	a.Set.Validators = nil
	return true
}

func clearResourceSetNestedValidators(a *resource.Attribute) bool {
	if a.SetNested == nil {
		return false
	}
	a.SetNested.Validators = nil
	return true
}

func clearResourceMapValidators(a *resource.Attribute) bool {
	if a.Map == nil {
		return false
	}
	a.Map.Validators = nil
	return true
}

func clearResourceMapNestedValidators(a *resource.Attribute) bool {
	if a.MapNested == nil {
		return false
	}
	a.MapNested.Validators = nil
	return true
}

func clearResourceObjectValidators(a *resource.Attribute) bool {
	if a.Object == nil {
		return false
	}
	a.Object.Validators = nil
	return true
}

func clearResourceSingleNestedValidators(a *resource.Attribute) bool {
	if a.SingleNested == nil {
		return false
	}
	a.SingleNested.Validators = nil
	return true
}

func clearDSValidators(a *datasource.Attribute) bool {
	clearers := []func(*datasource.Attribute) bool{
		clearDSBoolValidators,
		clearDSStringValidators,
		clearDSInt64Validators,
		clearDSFloat64Validators,
		clearDSNumberValidators,
		clearDSListValidators,
		clearDSListNestedValidators,
		clearDSSetValidators,
		clearDSSetNestedValidators,
		clearDSMapValidators,
		clearDSMapNestedValidators,
		clearDSObjectValidators,
		clearDSSingleNestedValidators,
	}
	for _, clear := range clearers {
		if clear(a) {
			return true
		}
	}
	return false
}

func clearDSBoolValidators(a *datasource.Attribute) bool {
	if a.Bool == nil {
		return false
	}
	a.Bool.Validators = nil
	return true
}

func clearDSStringValidators(a *datasource.Attribute) bool {
	if a.String == nil {
		return false
	}
	a.String.Validators = nil
	return true
}

func clearDSInt64Validators(a *datasource.Attribute) bool {
	if a.Int64 == nil {
		return false
	}
	a.Int64.Validators = nil
	return true
}

func clearDSFloat64Validators(a *datasource.Attribute) bool {
	if a.Float64 == nil {
		return false
	}
	a.Float64.Validators = nil
	return true
}

func clearDSNumberValidators(a *datasource.Attribute) bool {
	if a.Number == nil {
		return false
	}
	a.Number.Validators = nil
	return true
}

func clearDSListValidators(a *datasource.Attribute) bool {
	if a.List == nil {
		return false
	}
	a.List.Validators = nil
	return true
}

func clearDSListNestedValidators(a *datasource.Attribute) bool {
	if a.ListNested == nil {
		return false
	}
	a.ListNested.Validators = nil
	return true
}

func clearDSSetValidators(a *datasource.Attribute) bool {
	if a.Set == nil {
		return false
	}
	a.Set.Validators = nil
	return true
}

func clearDSSetNestedValidators(a *datasource.Attribute) bool {
	if a.SetNested == nil {
		return false
	}
	a.SetNested.Validators = nil
	return true
}

func clearDSMapValidators(a *datasource.Attribute) bool {
	if a.Map == nil {
		return false
	}
	a.Map.Validators = nil
	return true
}

func clearDSMapNestedValidators(a *datasource.Attribute) bool {
	if a.MapNested == nil {
		return false
	}
	a.MapNested.Validators = nil
	return true
}

func clearDSObjectValidators(a *datasource.Attribute) bool {
	if a.Object == nil {
		return false
	}
	a.Object.Validators = nil
	return true
}

func clearDSSingleNestedValidators(a *datasource.Attribute) bool {
	if a.SingleNested == nil {
		return false
	}
	a.SingleNested.Validators = nil
	return true
}

// ── descriptions ────────────────────────────────────────────────────

// injectDescriptions overrides an attribute's description on the resource (or,
// for an `items.`-prefixed path, the matching datasource leaf). It is the one
// place a description for an OAS-derived attribute is declared, used when the
// provider deliberately rewords a user-facing field that regen would otherwise
// reset to the OAS text.
func injectDescriptions(s *spec.Specification, res *resource.Resource, resName string, specs []descriptionSpec) {
	for _, d := range specs {
		if d.Path == "" {
			fail("description with empty path in resource %q", resName)
		}
		value := d.Value
		if strings.HasPrefix(d.Path, "items.") {
			applyToDSItems(s, resName, d.Path, func(a *datasource.Attribute) bool { return setDSDescription(a, value) }, "description", "description")
			continue
		}
		applyToResourceAttr(
			res,
			resName,
			d.Path,
			func(a *resource.Attribute) bool { return setResourceDescription(a, value) },
			"description",
			"description",
		)
	}
}

func setResourceDescription(a *resource.Attribute, v string) bool {
	setters := []func(*resource.Attribute, string) bool{
		setResourceBoolDescription,
		setResourceStringDescription,
		setResourceInt64Description,
		setResourceFloat64Description,
		setResourceNumberDescription,
		setResourceListDescription,
		setResourceListNestedDescription,
		setResourceSetDescription,
		setResourceSetNestedDescription,
		setResourceMapDescription,
		setResourceMapNestedDescription,
		setResourceObjectDescription,
		setResourceSingleNestedDescription,
	}
	for _, set := range setters {
		if set(a, v) {
			return true
		}
	}
	return false
}

func setResourceBoolDescription(a *resource.Attribute, v string) bool {
	if a.Bool == nil {
		return false
	}
	a.Bool.Description = &v
	return true
}

func setResourceStringDescription(a *resource.Attribute, v string) bool {
	if a.String == nil {
		return false
	}
	a.String.Description = &v
	return true
}

func setResourceInt64Description(a *resource.Attribute, v string) bool {
	if a.Int64 == nil {
		return false
	}
	a.Int64.Description = &v
	return true
}

func setResourceFloat64Description(a *resource.Attribute, v string) bool {
	if a.Float64 == nil {
		return false
	}
	a.Float64.Description = &v
	return true
}

func setResourceNumberDescription(a *resource.Attribute, v string) bool {
	if a.Number == nil {
		return false
	}
	a.Number.Description = &v
	return true
}

func setResourceListDescription(a *resource.Attribute, v string) bool {
	if a.List == nil {
		return false
	}
	a.List.Description = &v
	return true
}

func setResourceListNestedDescription(a *resource.Attribute, v string) bool {
	if a.ListNested == nil {
		return false
	}
	a.ListNested.Description = &v
	return true
}

func setResourceSetDescription(a *resource.Attribute, v string) bool {
	if a.Set == nil {
		return false
	}
	a.Set.Description = &v
	return true
}

func setResourceSetNestedDescription(a *resource.Attribute, v string) bool {
	if a.SetNested == nil {
		return false
	}
	a.SetNested.Description = &v
	return true
}

func setResourceMapDescription(a *resource.Attribute, v string) bool {
	if a.Map == nil {
		return false
	}
	a.Map.Description = &v
	return true
}

func setResourceMapNestedDescription(a *resource.Attribute, v string) bool {
	if a.MapNested == nil {
		return false
	}
	a.MapNested.Description = &v
	return true
}

func setResourceObjectDescription(a *resource.Attribute, v string) bool {
	if a.Object == nil {
		return false
	}
	a.Object.Description = &v
	return true
}

func setResourceSingleNestedDescription(a *resource.Attribute, v string) bool {
	if a.SingleNested == nil {
		return false
	}
	a.SingleNested.Description = &v
	return true
}

func setDSDescription(a *datasource.Attribute, v string) bool {
	setters := []func(*datasource.Attribute, string) bool{
		setDSBoolDescription,
		setDSStringDescription,
		setDSInt64Description,
		setDSFloat64Description,
		setDSNumberDescription,
		setDSListDescription,
		setDSListNestedDescription,
		setDSSetDescription,
		setDSSetNestedDescription,
		setDSMapDescription,
		setDSMapNestedDescription,
		setDSObjectDescription,
		setDSSingleNestedDescription,
	}
	for _, set := range setters {
		if set(a, v) {
			return true
		}
	}
	return false
}

func setDSBoolDescription(a *datasource.Attribute, v string) bool {
	if a.Bool == nil {
		return false
	}
	a.Bool.Description = &v
	return true
}

func setDSStringDescription(a *datasource.Attribute, v string) bool {
	if a.String == nil {
		return false
	}
	a.String.Description = &v
	return true
}

func setDSInt64Description(a *datasource.Attribute, v string) bool {
	if a.Int64 == nil {
		return false
	}
	a.Int64.Description = &v
	return true
}

func setDSFloat64Description(a *datasource.Attribute, v string) bool {
	if a.Float64 == nil {
		return false
	}
	a.Float64.Description = &v
	return true
}

func setDSNumberDescription(a *datasource.Attribute, v string) bool {
	if a.Number == nil {
		return false
	}
	a.Number.Description = &v
	return true
}

func setDSListDescription(a *datasource.Attribute, v string) bool {
	if a.List == nil {
		return false
	}
	a.List.Description = &v
	return true
}

func setDSListNestedDescription(a *datasource.Attribute, v string) bool {
	if a.ListNested == nil {
		return false
	}
	a.ListNested.Description = &v
	return true
}

func setDSSetDescription(a *datasource.Attribute, v string) bool {
	if a.Set == nil {
		return false
	}
	a.Set.Description = &v
	return true
}

func setDSSetNestedDescription(a *datasource.Attribute, v string) bool {
	if a.SetNested == nil {
		return false
	}
	a.SetNested.Description = &v
	return true
}

func setDSMapDescription(a *datasource.Attribute, v string) bool {
	if a.Map == nil {
		return false
	}
	a.Map.Description = &v
	return true
}

func setDSMapNestedDescription(a *datasource.Attribute, v string) bool {
	if a.MapNested == nil {
		return false
	}
	a.MapNested.Description = &v
	return true
}

func setDSObjectDescription(a *datasource.Attribute, v string) bool {
	if a.Object == nil {
		return false
	}
	a.Object.Description = &v
	return true
}

func setDSSingleNestedDescription(a *datasource.Attribute, v string) bool {
	if a.SingleNested == nil {
		return false
	}
	a.SingleNested.Description = &v
	return true
}

func toInt64(v any) (int64, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case float64:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("%w, got %T", errIntStaticMismatch, v)
	}
}

func toFloat64(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("%w, got %T", errFloatStaticMismatch, v)
	}
}

// ── shared ──────────────────────────────────────────────────────────

// resolveAttribute descends the dotted path segments, stepping into nested
// objects, and returns a pointer to the matching attribute (or nil).
func resolveAttribute(attrs resource.Attributes, segments []string) *resource.Attribute {
	if len(segments) == 0 {
		return nil
	}
	for i := range attrs {
		if attrs[i].Name != segments[0] {
			continue
		}
		if len(segments) == 1 {
			return &attrs[i]
		}
		child := childAttributes(&attrs[i])
		if child == nil {
			return nil
		}
		return resolveAttribute(child, segments[1:])
	}
	return nil
}

// childAttributes returns the nested attribute set for a nested attribute, or
// nil for a leaf.
func childAttributes(a *resource.Attribute) resource.Attributes {
	switch {
	case a.ListNested != nil:
		return a.ListNested.NestedObject.Attributes
	case a.SetNested != nil:
		return a.SetNested.NestedObject.Attributes
	case a.MapNested != nil:
		return a.MapNested.NestedObject.Attributes
	case a.SingleNested != nil:
		return a.SingleNested.Attributes
	}
	return nil
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "inject-schema-overrides: "+format+"\n", args...)
	os.Exit(1)
}

package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	// titleCaser title-cases ASCII words for snake_to_Pascal conversion.
	titleCaser = cases.Title(language.English)
	// errSDKCreateRequestBodyTypeNotFound is returned when the SDK create request
	// body type cannot be located.
	errSDKCreateRequestBodyTypeNotFound = errors.New("SDK create request body type not found")
	// errUnsupportedCollectionKind is returned for an unsupported nested field
	// collection kind.
	errUnsupportedCollectionKind = errors.New("unsupported collection kind for nested field")
	// errSDKTypeNotStructOrAlias is returned when an SDK type is neither a struct
	// nor an alias.
	errSDKTypeNotStructOrAlias = errors.New("SDK type is not a struct or alias")
	// errSDKTypeCircularAlias is returned when an SDK type alias chain loops.
	errSDKTypeCircularAlias = errors.New("SDK type has a circular alias")
	// errSDKAliasNotFound is returned when an SDK alias target is missing.
	errSDKAliasNotFound = errors.New("SDK alias not found")
	// errSDKStructNotFound is returned when an SDK struct type is missing.
	errSDKStructNotFound = errors.New("SDK struct not found")
	// errSDKTypeAmbiguous is returned when an SDK type name matches multiple
	// types.
	errSDKTypeAmbiguous = errors.New("SDK type is ambiguous")
	// errTFModelNotFound is returned when a Terraform model type is missing.
	errTFModelNotFound = errors.New("TF model not found")
)

// ---------- Public types ----------

type FieldMapping struct {
	TFFieldName     string
	TFType          string
	TFSdkTag        string
	SDKFieldName    string
	SDKType         string
	SerializeExpr   string
	DeserializeExpr string
	DatasourceExpr  string
	IsTODO          bool
	TODOReason      string
	// NullOnRead marks a TF leaf absent from the read SDK struct: the read
	// serializer emits a typed null (SerializeExpr/DatasourceExpr) and the read
	// deserializer skips it entirely. See Descriptor.NullOnRead.
	NullOnRead bool
	// Nested is set when this field requires a generated sub-serializer.
	Nested *NestedSerializer
}

// typedNullExpr returns the framework typed-null constructor for a scalar TF
// type, used by the null_on_read knob. Returns ("", false) for unsupported
// (e.g. collection/object) types, which the caller turns into a TODO.
func typedNullExpr(tfType string) (string, bool) {
	switch tfType {
	case "types.String":
		return "types.StringNull()", true
	case "types.Int64":
		return "types.Int64Null()", true
	case "types.Bool":
		return "types.BoolNull()", true
	case "types.Float64":
		return "types.Float64Null()", true
	case "types.Number":
		return "types.NumberNull()", true
	default:
		return "", false
	}
}

// NestedSerializer describes a generated sub-serializer for a nested SDK struct
// mapped to a TF ObjectValuable type.
type NestedSerializer struct {
	TFFieldName     string
	SDKFieldName    string
	SDKStructName   string
	TFValueTypeName string
	PackageAlias    string // e.g. resource_security_group or datasource_security_group
	CollectionKind  string // "set", "list" or "object"
	// ParentSDKStruct is the SDK struct that *owns* this field (the top-level read
	// struct for top-level fields, or the enclosing element struct for fields
	// nested inside another object). The datasource list-mapper takes this as its
	// `item` parameter so doubly-nested lists access the correct parent.
	ParentSDKStruct        string
	SDKPointer             bool // true when SDK field is *SomeStruct or *[]SomeStruct
	SerializeFuncName      string
	TFOwnerType            string // original TF field type (e.g. "types.Object", "PlacementValue")
	DeserializeFuncName    string
	DatasourceFuncName     string // helper for list/set of nested objects in datasources
	DatasourceItemFuncName string // per-item mapper for datasources
	Fields                 []FieldMapping
	// DeserializeNeedsCtx/Diags guide the generated deserialize helper signature.
	// For set/list helpers ctx is always required (ElementsAs) and diags is never
	// used. For object helpers the flags reflect whether the field expressions
	// reference ctx/diags; the types.Object variant always needs both.
	DeserializeNeedsCtx   bool
	DeserializeNeedsDiags bool
}

func (n *NestedSerializer) IsCollection() bool {
	return n.CollectionKind == collectionKindSet || n.CollectionKind == collectionKindList
}

func (n *NestedSerializer) IsObject() bool {
	return n.CollectionKind == collectionKindObject
}

type MappingResult struct {
	SerializeFields             []FieldMapping
	CreateResponseFields        []FieldMapping
	DeserializeFields           []FieldMapping
	DatasourceFields            []FieldMapping
	DatasourceParams            []FieldMapping
	NestedSerializers           []NestedSerializer // used by resource file
	NestedDatasourceSerializers []NestedSerializer // used by datasource file
	CreateRequestType           string             // resolved SDK create request body type name
}

// ---------- High-level API ----------

func BuildMappings(d *Descriptor, projectDir string) (*MappingResult, error) {
	res := &MappingResult{}

	if d.HasResource {
		if err := buildResourceSection(d, projectDir, res); err != nil {
			return nil, err
		}
	}

	// Deduplicate nested serializers when the same (SDK struct, Value type) pair
	// is used by multiple serializers (read + create response).
	res.NestedSerializers = dedupeNestedSerializers(res.NestedSerializers)

	if d.HasDatasource {
		if err := buildDatasourceSection(d, projectDir, res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func buildResourceSection(d *Descriptor, projectDir string, res *MappingResult) error {
	sdkReadPath := filepath.Join(projectDir, "internal", "sdk", "api", "numspot-client.go")
	tfModelPath := resourceGenPath(projectDir, d)
	tfPackageAlias := d.ResourceModelPackage()

	readMappings, err := buildResourceMappings(
		d,
		sdkReadPath,
		d.SDKSingularName(),
		tfModelPath,
		d.Singular+"Model",
		tfPackageAlias,
		mappingKindSerialize,
		d.MappingOverrides.Serialize,
		&res.NestedSerializers,
	)
	if err != nil {
		return fmt.Errorf("build read mappings: %w", err)
	}
	res.SerializeFields = readMappings

	if d.CreateResponseType != "" {
		createResponseMappings, err := buildResourceMappings(
			d,
			sdkReadPath,
			d.CreateResponseType,
			tfModelPath,
			d.Singular+"Model",
			tfPackageAlias,
			mappingKindCreateResponse,
			d.MappingOverrides.CreateResponse,
			&res.NestedSerializers,
		)
		if err != nil {
			return fmt.Errorf("build create response mappings: %w", err)
		}
		res.CreateResponseFields = createResponseMappings
	}

	// Resources whose create API takes no request body (configured purely via
	// follow-up link calls) skip create-body resolution and deserialization.
	if d.Core.Create.NoRequestBody {
		return nil
	}

	createTypeName, err := resolveCreateRequestBodyType(sdkReadPath, d)
	if err != nil {
		return fmt.Errorf("build create mappings: %w", err)
	}
	res.CreateRequestType = createTypeName

	createMappings, err := buildResourceMappings(
		d,
		sdkReadPath,
		createTypeName,
		tfModelPath,
		d.Singular+"Model",
		tfPackageAlias,
		mappingKindDeserialize,
		d.MappingOverrides.Deserialize,
		&res.NestedSerializers,
	)
	if err != nil {
		return fmt.Errorf("build create mappings: %w", err)
	}
	res.DeserializeFields = createMappings
	return nil
}

func buildDatasourceSection(d *Descriptor, projectDir string, res *MappingResult) error {
	sdkItemPath := filepath.Join(projectDir, "internal", "sdk", "api", "numspot-client.go")
	tfModelPath := datasourceGenPath(projectDir, d)
	tfPackageAlias := d.DatasourceModelPackage()

	dsMappings, err := buildResourceMappings(
		d,
		sdkItemPath,
		d.SDKSingularName(),
		tfModelPath,
		"ItemsValue",
		tfPackageAlias,
		mappingKindDatasource,
		d.MappingOverrides.Datasource,
		&res.NestedDatasourceSerializers,
	)
	if err != nil {
		return fmt.Errorf("build datasource item mappings: %w", err)
	}
	res.DatasourceFields = dsMappings
	res.NestedDatasourceSerializers = dedupeNestedSerializers(res.NestedDatasourceSerializers)

	// Parent-scoped lists take the parent id positionally (see
	// Descriptor.ParentKey), not an api.* params struct, so skip params
	// resolution entirely.
	if paramsType := d.CoreReadListParamsType(); paramsType != "" && !d.HasParentKey() {
		paramsMappings, err := buildDatasourceParamsMappings(
			sdkItemPath,
			paramsType,
			tfModelPath,
			d.Singular+"Model",
			d.MappingOverrides.DatasourceParams,
		)
		if err != nil {
			return fmt.Errorf("build datasource params mappings: %w", err)
		}
		res.DatasourceParams = paramsMappings
	}
	return nil
}

// ---------- Builders ----------

func resourceGenPath(projectDir string, d *Descriptor) string {
	snake := singularSnake(d.Singular)
	return filepath.Join(projectDir, "internal", "services", d.PackageName(), "resource_"+snake, snake+"_resource_gen.go")
}

func datasourceGenPath(projectDir string, d *Descriptor) string {
	snake := singularSnake(d.Singular)
	return filepath.Join(projectDir, "internal", "services", d.PackageName(), "datasource_"+snake, snake+"_data_source_gen.go")
}

func buildResourceMappings(
	d *Descriptor,
	sdkPath, sdkStructName, tfModelPath, tfModelName, tfPackageAlias, kind string,
	overrides []OverrideConfig,
	nested *[]NestedSerializer,
) ([]FieldMapping, error) {
	sdkFields, err := parseSDKStruct(sdkPath, sdkStructName)
	if err != nil {
		return nil, err
	}

	tfFields, err := parseTFModel(tfModelPath, tfModelName)
	if err != nil {
		return nil, err
	}

	return matchFields(d, projectDirFromPath(sdkPath), tfPackageAlias, kind, "", sdkStructName, tfFields, sdkFields, overrides, nested), nil
}

func buildDatasourceParamsMappings(sdkPath, sdkStructName, tfModelPath, tfModelName string, overrides []OverrideConfig) ([]FieldMapping, error) {
	sdkFields, err := parseSDKStruct(sdkPath, sdkStructName)
	if err != nil {
		return nil, err
	}

	tfFields, err := parseTFModel(tfModelPath, tfModelName)
	if err != nil {
		return nil, err
	}

	return matchFieldsReverse(tfFields, sdkFields, overrides), nil
}

func createRequestBodyTypeName(singular string) string {
	return "Create" + singular + "JSONRequestBody"
}

// resolveCreateRequestBodyType returns the SDK request body type for the create
// call. Some APIs use the plural form (CreateVmsJSONRequestBody) while others
// use the singular form (CreateVpcJSONRequestBody), so both are attempted.
func resolveCreateRequestBodyType(sdkReadPath string, d *Descriptor) (string, error) {
	// An explicit override wins: resolve it for correct casing, but otherwise use
	// it verbatim (the create body does not follow the Create*JSONRequestBody
	// naming pattern).
	if d.Core.Create.RequestBodyType != "" {
		if resolved, err := resolveSDKStructName(sdkReadPath, d.Core.Create.RequestBodyType); err == nil {
			return resolved, nil
		}
		return d.Core.Create.RequestBodyType, nil
	}
	candidates := []string{
		createRequestBodyTypeName(d.Singular),
		createRequestBodyTypeName(d.Plural),
		createRequestBodyTypeName(d.SDKSingularName()),
		createRequestBodyTypeName(d.SDKPluralName()),
	}
	for _, candidate := range candidates {
		// Return the *resolved* struct name (correct casing), not the candidate:
		// parseSDKStruct matches case/separator-insensitively, so a candidate of
		// "CreateVpnConnectionJSONRequestBody" resolves to the real
		// "CreateVPNConnectionJSONRequestBody", which is what the generated code
		// must reference.
		if resolved, err := resolveSDKStructName(sdkReadPath, candidate); err == nil {
			return resolved, nil
		}
	}
	return "", fmt.Errorf("%w for %s or %s", errSDKCreateRequestBodyTypeNotFound, d.Singular, d.Plural)
}

func projectDirFromPath(sdkPath string) string {
	// sdkPath is expected to be <projectDir>/internal/sdk/api/numspot-client.go
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(sdkPath))))
}

// ---------- Field matching ----------

type sdkField struct {
	Name    string
	Type    string
	JSONTag string
}

type tfField struct {
	Name     string
	Type     string
	TFSdkTag string
}

func matchFields(
	d *Descriptor,
	projectDir, tfPackageAlias, kind, pathPrefix, parentSDKStruct string,
	tfFields []tfField,
	sdkFields []sdkField,
	overrides []OverrideConfig,
	nested *[]NestedSerializer,
) []FieldMapping {
	sdkByName, sdkByNormalized, normalizedDup := buildSDKLookupMaps(sdkFields)
	overrideByTF, overrideByPath := buildOverrideMaps(overrides)

	var mappings []FieldMapping
	for _, tf := range tfFields {
		fullPath := tfFieldPath(pathPrefix, tf.TFSdkTag)
		over, hasOverride := findOverride(overrideByTF, overrideByPath, tf.Name, fullPath)
		if hasOverride && over.Skip {
			continue
		}
		if tf.TFSdkTag == "tags" {
			continue
		}
		if shouldSkipByPattern(d, kind, tf, hasOverride, &over) {
			continue
		}

		candidate := resolveCandidateName(d, tf, hasOverride, &over)
		sdk, found := lookupSDKField(candidate, sdkByName, sdkByNormalized, normalizedDup)

		mapping, skip := buildMapping(
			d,
			projectDir,
			tfPackageAlias,
			kind,
			fullPath,
			parentSDKStruct,
			tf,
			sdk,
			found,
			hasOverride,
			&over,
			overrides,
			nested,
		)
		if skip {
			continue
		}
		mappings = append(mappings, mapping)
	}

	return mappings
}

func buildSDKLookupMaps(sdkFields []sdkField) (byName, byNormalized map[string]sdkField, normalizedDup map[string]bool) {
	byName = make(map[string]sdkField, len(sdkFields))
	byNormalized = make(map[string]sdkField, len(sdkFields))
	normalizedDup = make(map[string]bool, len(sdkFields))
	for _, f := range sdkFields {
		byName[f.Name] = f
		key := normalizeSDKName(f.Name)
		if _, exists := byNormalized[key]; exists {
			normalizedDup[key] = true
		}
		byNormalized[key] = f
	}
	return
}

func buildOverrideMaps(overrides []OverrideConfig) (byTF, byPath map[string]OverrideConfig) {
	byTF = make(map[string]OverrideConfig, len(overrides))
	byPath = make(map[string]OverrideConfig, len(overrides))
	for _, o := range overrides {
		if o.Path != "" {
			byPath[o.Path] = o
		} else {
			byTF[o.TFField] = o
		}
	}
	return
}

func tfFieldPath(pathPrefix, tag string) string {
	if pathPrefix == "" {
		return tag
	}
	return pathPrefix + "." + tag
}

func findOverride(byTF, byPath map[string]OverrideConfig, tfName, fullPath string) (OverrideConfig, bool) {
	if o, ok := byPath[fullPath]; ok {
		return o, true
	}
	if o, ok := byTF[tfName]; ok {
		return o, true
	}
	return OverrideConfig{}, false
}

func shouldSkipByPattern(d *Descriptor, kind string, tf tfField, hasOverride bool, over *OverrideConfig) bool {
	switch kind {
	case mappingKindSerialize, mappingKindDatasource:
		// Preserved fields are not derived from the read/list payload;
		// the resource methods copy them from prior state instead.
		// An explicit expression override wins over the skip: it opts the
		// field back into the generated mapping (e.g. a preserved resource
		// input that the datasource still maps live from the list payload).
		if d.IsPreserveFromState(tf.Name) && (!hasOverride || over.Expression == "") {
			return true
		}
	case mappingKindDeserialize:
		if d.IsExcludeFromRequest(tf.Name) || d.IsComputedField(tf.Name) {
			return true
		}
	}
	return false
}

func resolveCandidateName(d *Descriptor, tf tfField, hasOverride bool, over *OverrideConfig) string {
	candidate := snakeToPascal(tf.TFSdkTag)
	// A renamed nested leaf (see descriptor `renames`) no longer matches its
	// SDK field by name; recover the original OAS leaf so it does.
	if orig, ok := d.OriginalLeafForRenamed(tf.TFSdkTag); ok {
		candidate = snakeToPascal(orig)
	}
	if d.IsIDSource(tf.Name) {
		return d.IDSource
	}
	if hasOverride && over.SDKField != "" {
		return over.SDKField
	}
	return candidate
}

func lookupSDKField(candidate string, byName, byNormalized map[string]sdkField, normalizedDup map[string]bool) (sdkField, bool) {
	if sdk, ok := byName[candidate]; ok {
		return sdk, true
	}
	key := normalizeSDKName(candidate)
	if normalizedDup[key] {
		return sdkField{}, false
	}
	sdk, ok := byNormalized[key]
	return sdk, ok
}

func buildMapping(
	d *Descriptor,
	projectDir, tfPackageAlias, kind, fullPath, parentSDKStruct string,
	tf tfField,
	sdk sdkField,
	found, hasOverride bool,
	over *OverrideConfig,
	overrides []OverrideConfig,
	nested *[]NestedSerializer,
) (FieldMapping, bool) {
	mapping := FieldMapping{
		TFFieldName:  tf.Name,
		TFType:       tf.Type,
		TFSdkTag:     tf.TFSdkTag,
		SDKFieldName: sdk.Name,
		SDKType:      sdk.Type,
	}

	if hasOverride && over.Expression != "" {
		// An explicit expression override always wins, even for derived
		// fields that have no matching SDK field (e.g. a value computed
		// from a nested list). This is the escape hatch for mappings the
		// generator cannot infer.
		mapping.SerializeExpr = over.Expression
		mapping.DeserializeExpr = over.Expression
		mapping.DatasourceExpr = over.Expression
		return mapping, false
	}

	if !found {
		// A preserved field that the create response does not carry is
		// sourced from the plan at the call site, so skip it here rather
		// than emitting a TODO.
		if kind == mappingKindCreateResponse && d.IsPreserveFromState(tf.Name) {
			return FieldMapping{}, true
		}
		// A null_on_read leaf is present in the TF schema but absent from the
		// read SDK struct: serialize a typed null and let the deserializer
		// skip it (see GenerateNested(De)serializeBody).
		if nullExpr, ok := typedNullExpr(tf.Type); ok && d.IsNullOnRead(fullPath) {
			mapping.NullOnRead = true
			mapping.SerializeExpr = nullExpr
			mapping.DatasourceExpr = nullExpr
		} else {
			mapping.IsTODO = true
			mapping.TODOReason = fmt.Sprintf("no SDK field matching tfsdk tag %q", tf.TFSdkTag)
		}
		return mapping, false
	}

	if resolved := tryResolveStruct(d, projectDir, tfPackageAlias, fullPath, parentSDKStruct, tf, sdk, overrides, nested); resolved != nil {
		return *resolved, false
	}

	resolve(&mapping)
	return mapping, false
}

func matchFieldsReverse(tfFields []tfField, sdkFields []sdkField, overrides []OverrideConfig) []FieldMapping {
	overrideByTF := make(map[string]OverrideConfig, len(overrides))
	for _, o := range overrides {
		overrideByTF[o.TFField] = o
	}

	tfByName := make(map[string]tfField, len(tfFields))
	for _, f := range tfFields {
		tfByName[f.Name] = f
	}

	var mappings []FieldMapping
	for _, sdk := range sdkFields {
		tf, found := tfByName[sdk.Name]
		if !found {
			continue
		}

		over, hasOverride := overrideByTF[tf.Name]
		if hasOverride && over.Skip {
			continue
		}

		mapping := FieldMapping{
			TFFieldName:  tf.Name,
			TFType:       tf.Type,
			TFSdkTag:     tf.TFSdkTag,
			SDKFieldName: sdk.Name,
			SDKType:      sdk.Type,
		}

		if hasOverride && over.Expression != "" {
			mapping.DatasourceExpr = over.Expression
		} else {
			resolveDatasourceParams(&mapping)
		}

		mappings = append(mappings, mapping)
	}

	return mappings
}

// ---------- Struct / nested resolution ----------

// tryResolveStruct detects SDK struct / slice-of-struct fields and builds the
// corresponding sub-serializer. It returns nil when the field is a simple type.
func tryResolveStruct(
	d *Descriptor,
	projectDir, tfPackageAlias, pathPrefix, parentSDKStruct string,
	tf tfField,
	sdk sdkField,
	overrides []OverrideConfig,
	nested *[]NestedSerializer,
) *FieldMapping {
	sdkType := sdk.Type
	tfType := tf.Type

	isPtr := false
	if strings.HasPrefix(sdkType, "*") {
		isPtr = true
		sdkType = strings.TrimPrefix(sdkType, "*")
	}

	isSlice := strings.HasPrefix(sdkType, "[]")

	collectionKind := inferCollectionKind(tfType)
	if collectionKind == "" {
		return nil
	}

	// Single nested objects cannot be mapped from a slice.
	if collectionKind == collectionKindObject && isSlice {
		return nil
	}

	tfValueTypeName := resolveTFValueTypeName(d, tf, tfType, collectionKind)
	tfModelPath := tfModelPathForPackage(projectDir, tfPackageAlias, d)
	if _, err := parseTFModel(tfModelPath, tfValueTypeName); err != nil {
		return nil
	}

	// Build the nested serializer. If it fails, fall back to TODO.
	ns, err := buildNestedSerializer(
		d,
		projectDir,
		tfPackageAlias,
		pathPrefix,
		parentSDKStruct,
		tf,
		sdk,
		tfValueTypeName,
		collectionKind,
		isPtr,
		isSlice,
		overrides,
		nested,
	)
	if err != nil {
		return &FieldMapping{
			TFFieldName:  tf.Name,
			TFType:       tf.Type,
			TFSdkTag:     tf.TFSdkTag,
			SDKFieldName: sdk.Name,
			SDKType:      sdk.Type,
			IsTODO:       true,
			TODOReason:   fmt.Sprintf("nested mapping failed for %s: %v", tf.Name, err),
		}
	}

	*nested = append(*nested, *ns)

	m := &FieldMapping{
		TFFieldName:  tf.Name,
		TFType:       tf.Type,
		TFSdkTag:     tf.TFSdkTag,
		SDKFieldName: sdk.Name,
		SDKType:      sdk.Type,
		Nested:       ns,
	}
	buildNestedExpressions(m, ns, collectionKind, isPtr, sdk)
	return m
}

func inferCollectionKind(tfType string) string {
	switch tfType {
	case "types.List":
		return collectionKindList
	case "types.Set":
		return collectionKindSet
	case tfTypeObject:
		return collectionKindObject
	default:
		// The terraform-plugin-framework generator often uses a custom
		// ObjectValuable type named <Field>Value for single nested objects.
		if strings.HasSuffix(tfType, "Value") {
			return collectionKindObject
		}
	}
	return ""
}

func resolveTFValueTypeName(d *Descriptor, tf tfField, tfType, collectionKind string) string {
	// The TF ObjectValuable type name is the TF field name + "Value"
	// (e.g. PublicIps -> PublicIpsValue), which is how the framework generator
	// names both single-object and list/set element value types. Only when the
	// field is *itself* a custom ObjectValuable type (collectionKind "object"
	// derived from a "<X>Value" type rather than types.Object) do we trust the
	// declared type directly. For types.List/types.Set the declared type is the
	// generic collection, so we must keep the derived <Field>Value name.
	tfValueTypeName := d.NestedTypeName(tf.Name)
	if collectionKind == collectionKindObject && strings.HasSuffix(tfType, "Value") {
		tfValueTypeName = tfType
	}
	return tfValueTypeName
}

func tfModelPathForPackage(projectDir, tfPackageAlias string, d *Descriptor) string {
	if strings.HasPrefix(tfPackageAlias, "datasource_") {
		return datasourceGenPath(projectDir, d)
	}
	return resourceGenPath(projectDir, d)
}

func buildNestedExpressions(m *FieldMapping, ns *NestedSerializer, collectionKind string, isPtr bool, sdk sdkField) {
	switch collectionKind {
	case collectionKindSet:
		m.SerializeExpr = genericSetSerializeExpr(ns, sdk, isPtr)
		m.DeserializeExpr = buildNestedDeserializeCall(ns.DeserializeFuncName, m.TFFieldName, ns.DeserializeNeedsCtx, ns.DeserializeNeedsDiags)
		m.DatasourceExpr = fmt.Sprintf("mapping%s(ctx, item, &diags)", m.TFFieldName)
	case collectionKindList:
		m.SerializeExpr = genericListSerializeExpr(ns, sdk, isPtr)
		m.DeserializeExpr = buildNestedDeserializeCall(ns.DeserializeFuncName, m.TFFieldName, ns.DeserializeNeedsCtx, ns.DeserializeNeedsDiags)
		m.DatasourceExpr = fmt.Sprintf("mapping%s(ctx, item, &diags)", m.TFFieldName)
	case collectionKindObject:
		m.SerializeExpr = objectSerializeExpr(ns, sdk, isPtr)
		m.DeserializeExpr = buildNestedDeserializeCall(ns.DeserializeFuncName, m.TFFieldName, ns.DeserializeNeedsCtx, ns.DeserializeNeedsDiags)
		m.DatasourceExpr = objectDatasourceExpr(ns, sdk, isPtr)
	}

	// SDK create request bodies often use *[]T; the nested deserializer returns a
	// value slice ([]T), so wrap it to get the pointer. (&call(...) is not valid
	// Go, so this must go through a helper.)
	if isPtr && (collectionKind == collectionKindSet || collectionKind == collectionKindList) {
		m.DeserializeExpr = fmt.Sprintf("utils.PointerOf(%s)", m.DeserializeExpr)
	}
}

func genericSetSerializeExpr(ns *NestedSerializer, sdk sdkField, isPtr bool) string {
	// Deref only a pointer-to-slice (*[]T); a value slice ([]T) is passed
	// directly. GenericSetToTfSetValue takes a value slice.
	if isPtr {
		return fmt.Sprintf("utils.GenericSetToTfSetValue(ctx, %s, *http.%s, diags)", ns.SerializeFuncName, sdk.Name)
	}
	return fmt.Sprintf("utils.GenericSetToTfSetValue(ctx, %s, http.%s, diags)", ns.SerializeFuncName, sdk.Name)
}

func genericListSerializeExpr(ns *NestedSerializer, sdk sdkField, isPtr bool) string {
	// Deref only a pointer-to-slice (*[]T); a value slice ([]T) is passed
	// directly. GenericListToTfListValue takes a value slice.
	if isPtr {
		return fmt.Sprintf("utils.GenericListToTfListValue(ctx, %s, *http.%s, diags)", ns.SerializeFuncName, sdk.Name)
	}
	return fmt.Sprintf("utils.GenericListToTfListValue(ctx, %s, http.%s, diags)", ns.SerializeFuncName, sdk.Name)
}

func objectSerializeExpr(ns *NestedSerializer, sdk sdkField, isPtr bool) string {
	// Sub-serializer accepts a pointer so nil SDK values can be represented.
	if isPtr {
		return fmt.Sprintf("%s(ctx, http.%s, diags)", ns.SerializeFuncName, sdk.Name)
	}
	return fmt.Sprintf("%s(ctx, &http.%s, diags)", ns.SerializeFuncName, sdk.Name)
}

func objectDatasourceExpr(ns *NestedSerializer, sdk sdkField, isPtr bool) string {
	if isPtr {
		return fmt.Sprintf("%s(ctx, item.%s, &diags)", ns.DatasourceItemFuncName, sdk.Name)
	}
	return fmt.Sprintf("%s(ctx, &item.%s, &diags)", ns.DatasourceItemFuncName, sdk.Name)
}

func buildNestedSerializer(
	d *Descriptor,
	projectDir, tfPackageAlias, pathPrefix, parentSDKStruct string,
	tf tfField,
	sdk sdkField,
	tfValueTypeName, collectionKind string,
	sdkPtr, sdkSlice bool,
	overrides []OverrideConfig,
	nested *[]NestedSerializer,
) (*NestedSerializer, error) {
	sdkType := sdk.Type
	if sdkPtr {
		sdkType = strings.TrimPrefix(sdkType, "*")
	}
	if sdkSlice {
		sdkType = strings.TrimPrefix(sdkType, "[]")
	}

	if collectionKind == "" {
		return nil, fmt.Errorf("%w %s", errUnsupportedCollectionKind, tf.Name)
	}

	tfFields, sdkFields, err := loadNestedModelFields(projectDir, tfPackageAlias, tfValueTypeName, sdkType, d)
	if err != nil {
		return nil, err
	}

	// Thread the real accumulator so deeper nested objects (e.g.
	// vpn_options -> phase1options/phase2options) register their own
	// sub-serializers and get emitted, rather than being discarded. The element
	// struct (sdkType) is the parent of these sub-fields, so deeper nested list
	// mappers take it (not the top-level read struct) as their `item`.
	// Overrides from the current bucket are threaded too, so nested enum fields
	// (e.g. KubernetesStatusState) can be mapped via mapping_overrides.
	fields := matchFields(d, projectDir, tfPackageAlias, "nested", pathPrefix, sdkType, tfFields, sdkFields, overrides, nested)
	reStampExpressionOverrides(d, pathPrefix, fields)
	rewriteNestedFieldExpressions(fields)
	if strings.HasPrefix(tfPackageAlias, "datasource_") {
		adaptNestedDatasourceExpressions(fields, collectionKind == collectionKindObject)
	}

	ns := &NestedSerializer{
		TFFieldName:            tf.Name,
		SDKFieldName:           sdk.Name,
		SDKStructName:          sdkType,
		TFValueTypeName:        tfValueTypeName,
		PackageAlias:           tfPackageAlias,
		CollectionKind:         collectionKind,
		ParentSDKStruct:        parentSDKStruct,
		SDKPointer:             sdkPtr,
		TFOwnerType:            tf.Type,
		SerializeFuncName:      "serialize" + singularize(tf.Name),
		DeserializeFuncName:    "deserialize" + tf.Name,
		DatasourceFuncName:     "mapping" + tf.Name,
		DatasourceItemFuncName: "mapping" + singularize(tf.Name),
		Fields:                 fields,
	}
	configureNestedDeserializeNeeds(ns, tf.Type, fields)
	return ns, nil
}

func loadNestedModelFields(projectDir, tfPackageAlias, tfValueTypeName, sdkType string, d *Descriptor) ([]tfField, []sdkField, error) {
	tfModelPath := tfModelPathForPackage(projectDir, tfPackageAlias, d)
	tfFields, err := parseTFModel(tfModelPath, tfValueTypeName)
	if err != nil {
		return nil, nil, fmt.Errorf("parse TF value type %q: %w", tfValueTypeName, err)
	}

	sdkPath := filepath.Join(projectDir, "internal", "sdk", "api", "numspot-client.go")
	sdkFields, err := parseSDKStruct(sdkPath, sdkType)
	if err != nil {
		return nil, nil, fmt.Errorf("parse SDK struct %q: %w", sdkType, err)
	}
	return tfFields, sdkFields, nil
}

func reStampExpressionOverrides(d *Descriptor, pathPrefix string, fields []FieldMapping) {
	// An explicit expression override stamps all three expression slots
	// (matchFields has no direction context), but a nested serializer is built
	// once and shared across directions: without a re-stamp, whichever pass
	// builds it first (the read pass) would leave its expressions on the
	// deserialize/datasource slots too. Re-apply each bucket to its own slot.
	// Path-based overrides take precedence over name-based ones so that
	// same-named top-level and nested fields can be mapped independently.
	reStamp(d.MappingOverrides.Serialize, func(f *FieldMapping, e string) { f.SerializeExpr = e }, pathPrefix, fields)
	reStamp(d.MappingOverrides.CreateResponse, func(f *FieldMapping, e string) { f.SerializeExpr = e }, pathPrefix, fields)
	reStamp(d.MappingOverrides.Deserialize, func(f *FieldMapping, e string) { f.DeserializeExpr = e }, pathPrefix, fields)
	reStamp(d.MappingOverrides.Datasource, func(f *FieldMapping, e string) { f.DatasourceExpr = e }, pathPrefix, fields)
}

func reStamp(overrides []OverrideConfig, set func(*FieldMapping, string), pathPrefix string, fields []FieldMapping) {
	byPath, byTF := buildExpressionOverrideMaps(overrides)
	for i := range fields {
		fieldPath := fields[i].TFSdkTag
		if pathPrefix != "" {
			fieldPath = pathPrefix + "." + fields[i].TFSdkTag
		}
		if o, ok := byPath[fieldPath]; ok {
			set(&fields[i], o.Expression)
		} else if o, ok := byTF[fields[i].TFFieldName]; ok {
			set(&fields[i], o.Expression)
		}
	}
}

func buildExpressionOverrideMaps(overrides []OverrideConfig) (byPath, byTF map[string]OverrideConfig) {
	byPath = make(map[string]OverrideConfig, len(overrides))
	byTF = make(map[string]OverrideConfig, len(overrides))
	for _, o := range overrides {
		if o.Expression == "" {
			continue
		}
		if o.Path != "" {
			byPath[o.Path] = o
		} else {
			byTF[o.TFField] = o
		}
	}
	return
}

func rewriteNestedFieldExpressions(fields []FieldMapping) {
	// Rewrite generated expressions to reference the local function parameters
	// (the SDK item is named "item", the deserialized TF element is "e").
	for i := range fields {
		fields[i].SerializeExpr = strings.ReplaceAll(fields[i].SerializeExpr, "http.", "item.")
		fields[i].DeserializeExpr = strings.ReplaceAll(fields[i].DeserializeExpr, "tf.", "e.")
	}
}

func adaptNestedDatasourceExpressions(fields []FieldMapping, isObject bool) {
	// Datasource per-item mappers come in two shapes:
	//   - list/set element mappers use a *value* `diags diag.Diagnostics`, so
	//     helpers expecting *diag.Diagnostics must be passed `&diags`.
	//   - object mappers receive `diags *diag.Diagnostics` (a pointer), so the
	//     same helpers must be passed `diags` directly.
	// Additionally, an object/list/set sub-field nested inside a datasource
	// mapper must call the datasource mapper of the inner type (the "mapping..."
	// function), not the resource serializer ("serialize..."); the field's
	// DatasourceExpr already encodes that call shape, so reuse it here.
	for i := range fields {
		if fields[i].Nested != nil {
			fields[i].SerializeExpr = fields[i].DatasourceExpr
		}
		if isObject {
			fields[i].SerializeExpr = strings.ReplaceAll(fields[i].SerializeExpr, ", &diags)", ", diags)")
		} else {
			fields[i].SerializeExpr = strings.ReplaceAll(fields[i].SerializeExpr, ", diags)", ", &diags)")
		}
	}
}

func configureNestedDeserializeNeeds(ns *NestedSerializer, tfType string, fields []FieldMapping) {
	switch ns.CollectionKind {
	case collectionKindSet, collectionKindList:
		// ElementsAs always needs ctx; diags is required only when a nested field
		// expression (e.g. a nested object helper) references it.
		ns.DeserializeNeedsCtx = true
		ns.DeserializeNeedsDiags = nestedDeserializeBodyNeedsDiags(fields)
	case collectionKindObject:
		// The types.Object variant always calls obj.As(ctx, ..., diags).
		if tfType == tfTypeObject {
			ns.DeserializeNeedsCtx = true
			ns.DeserializeNeedsDiags = true
		} else {
			ns.DeserializeNeedsCtx, ns.DeserializeNeedsDiags = nestedDeserializeBodyNeedsCtxDiags(fields)
		}
	}
}

// singularize returns a naive singular form of a plural name.
// It strips a trailing 's' (but not 'ss'). This is enough for the common
// "Rules" -> "Rule" pattern. Irregular plurals can be overridden via
// nested_type_overrides if needed.
func singularize(s string) string {
	if strings.HasSuffix(s, "ss") {
		return s
	}
	if strings.HasSuffix(s, "s") {
		return strings.TrimSuffix(s, "s")
	}
	return s
}

// ---------- Type-pair resolution ----------

type exprTemplates struct {
	serialize, deserialize, datasource string
}

func applyTemplates(m *FieldMapping, t exprTemplates) {
	m.SerializeExpr = fmt.Sprintf(t.serialize, m.SDKFieldName)
	m.DeserializeExpr = fmt.Sprintf(t.deserialize, m.TFFieldName)
	m.DatasourceExpr = fmt.Sprintf(t.datasource, m.SDKFieldName)
}

//nolint:goconst // repeated format-string fragments are intentional template literals
var typePairTemplates = map[string]exprTemplates{
	"*string -> types.String": {
		"types.StringPointerValue(http.%s)",
		"tf.%s.ValueStringPointer()",
		"types.StringValue(utils.ConvertStringPtrToString(item.%s))",
	},
	"string -> types.String": {"types.StringValue(http.%s)", "tf.%s.ValueString()", "types.StringValue(item.%s)"},
	"*bool -> types.Bool":    {"types.BoolPointerValue(http.%s)", "tf.%s.ValueBoolPointer()", "types.BoolValue(*item.%s)"},
	"bool -> types.Bool":     {"types.BoolValue(http.%s)", "tf.%s.ValueBool()", "types.BoolValue(item.%s)"},
	"*int -> types.Int64": {
		"utils.FromIntPtrToTfInt64(http.%s)",
		"utils.FromTfInt64ToIntPtr(tf.%s)",
		"utils.FromIntPtrToTfInt64(item.%s)",
	},
	"int -> types.Int64": {
		"utils.FromIntToTfInt64(http.%s)",
		"utils.FromTfInt64ToInt(tf.%s)",
		"utils.FromIntToTfInt64(item.%s)",
	},
	"*AvailabilityZoneName -> types.String": {
		"types.StringValue(utils.ConvertAzNamePtrToString(http.%s))",
		"utils.FromTfStringToAzNamePtr(tf.%s)",
		"types.StringValue(utils.ConvertAzNamePtrToString(item.%s))",
	},
	"AvailabilityZoneName -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.AvailabilityZoneName(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"NodeProfiles -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.NodeProfiles(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"StatusState -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.StatusState(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"KubernetesNodePoolDiskType -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.KubernetesNodePoolDiskType(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"PostgreSQLClusterVisibility -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.PostgreSQLClusterVisibility(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"PostgreSQLStatusState -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.PostgreSQLStatusState(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"Visibility -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.Visibility(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"KubernetesClusterCreateProfile -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.KubernetesClusterCreateProfile(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"KubernetesClusterCreateVersion -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.KubernetesClusterCreateVersion(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"CreateVmsType -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.CreateVmsType(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"CreateFlexibleGpuModelName -> types.String": {
		"types.StringValue(string(http.%s))",
		"api.CreateFlexibleGpuModelName(tf.%s.ValueString())",
		"types.StringValue(string(item.%s))",
	},
	"*PostgreSQLMajorVersion -> types.String": {
		"types.StringPointerValue((*string)(http.%s))",
		"utils.PointerOf(api.PostgreSQLMajorVersion(tf.%s.ValueString()))",
		"types.StringPointerValue((*string)(item.%s))",
	},
	"*CreateFlexibleGpuGeneration -> types.String": {
		"types.StringPointerValue((*string)(http.%s))",
		"utils.PointerOf(api.CreateFlexibleGpuGeneration(tf.%s.ValueString()))",
		"types.StringPointerValue((*string)(item.%s))",
	},
	"PostgreSQLClusterName -> types.String": {"types.StringValue(http.%s)", "tf.%s.ValueString()", "types.StringValue(item.%s)"},
	"PostgreSQLUsername -> types.String":    {"types.StringValue(http.%s)", "tf.%s.ValueString()", "types.StringValue(item.%s)"},
	"*PostgreSQLFullVersion -> types.String": {
		"types.StringPointerValue(http.%s)",
		"tf.%s.ValueStringPointer()",
		"types.StringValue(utils.ConvertStringPtrToString(item.%s))",
	},
	"PostgreSQLClusterId -> types.String": {
		"types.StringValue(http.%s.String())",
		"utils.FromTfStringToUuid(tf.%s)",
		"types.StringValue(item.%s.String())",
	},
	"int32 -> types.Int64": {
		"types.Int64Value(int64(http.%s))",
		"int32(tf.%s.ValueInt64())",
		"types.Int64Value(int64(item.%s))",
	},
	"PostgreSQLMemorySizeGiB -> types.Int64": {
		"types.Int64Value(int64(http.%s))",
		"int32(tf.%s.ValueInt64())",
		"types.Int64Value(int64(item.%s))",
	},
	"PostgreSQLVCPUCount -> types.Int64": {
		"types.Int64Value(int64(http.%s))",
		"int32(tf.%s.ValueInt64())",
		"types.Int64Value(int64(item.%s))",
	},
	"PostgreSQLReplicaCount -> types.Int64": {
		"types.Int64Value(int64(http.%s))",
		"api.PostgreSQLReplicaCount(tf.%s.ValueInt64())",
		"types.Int64Value(int64(item.%s))",
	},
	"*int32 -> types.Int64": {
		"utils.FromInt32PtrToTfInt64(http.%s)",
		"utils.FromTfInt64ToInt32Ptr(tf.%s)",
		"utils.FromInt32PtrToTfInt64(item.%s)",
	},
	"*PostgreSQLReplicaCount -> types.Int64": {
		"utils.FromInt32PtrToTfInt64((*int32)(http.%s))",
		"utils.PointerOf(api.PostgreSQLReplicaCount(tf.%s.ValueInt64()))",
		"utils.FromInt32PtrToTfInt64((*int32)(item.%s))",
	},
	"StrictSlugMax63 -> types.String": {"types.StringValue(http.%s)", "tf.%s.ValueString()", "types.StringValue(item.%s)"},
	"*StrictSlugMax63 -> types.String": {
		"types.StringPointerValue(http.%s)",
		"tf.%s.ValueStringPointer()",
		"types.StringValue(utils.ConvertStringPtrToString(item.%s))",
	},
	"openapi_types.UUID -> types.String": {
		"types.StringValue(http.%s.String())",
		"utils.FromTfStringToUuid(tf.%s)",
		"types.StringValue(item.%s.String())",
	},
	"*openapi_types.UUID -> types.String": {
		"types.StringValue(utils.ConvertUuidPtrToString(http.%s))",
		"utils.FromTfStringToUuidPtr(tf.%s)",
		"types.StringValue(utils.ConvertUuidPtrToString(item.%s))",
	},
	"Uuid -> types.String": {
		"types.StringValue(http.%s.String())",
		"utils.FromTfStringToUuid(tf.%s)",
		"types.StringValue(item.%s.String())",
	},
	"*Uuid -> types.String": {
		"types.StringValue(utils.ConvertUuidPtrToString(http.%s))",
		"utils.FromTfStringToUuidPtr(tf.%s)",
		"types.StringValue(utils.ConvertUuidPtrToString(item.%s))",
	},
	"*time.Time -> types.String": {
		"utils.TimePtrToTfString(http.%s)",
		"utils.TfStringToTimePtr(tf.%s)",
		"utils.TimePtrToTfString(item.%s)",
	},
	"time.Time -> types.String": {
		"utils.TimeToTfString(http.%s)",
		"utils.TfStringToTime(tf.%s)",
		"utils.TimeToTfString(item.%s)",
	},
	"Cidr -> types.String": {"types.StringValue(http.%s)", "tf.%s.ValueString()", "types.StringValue(item.%s)"},
	"*Cidr -> types.String": {
		"types.StringPointerValue(http.%s)",
		"tf.%s.ValueStringPointer()",
		"types.StringValue(utils.ConvertStringPtrToString(item.%s))",
	},
	"*[]string -> types.List": {
		"utils.FromStringListPointerToTfStringList(ctx, http.%s, diags)",
		"utils.TfStringListToStringPtrList(ctx, tf.%s, diags)",
		"utils.FromStringListPointerToTfStringList(ctx, item.%s, &diags)",
	},
	"*[]string -> types.Set": {
		"utils.FromStringListPointerToTfStringSet(ctx, http.%s, diags)",
		"utils.FromTfStringSetToStringListPtr(ctx, tf.%s, diags)",
		"utils.FromStringListPointerToTfStringSet(ctx, item.%s, &diags)",
	},
	"[]string -> types.List": {
		"utils.FromStringListToTfStringList(ctx, http.%s, diags)",
		"utils.TfStringListToStringList(ctx, tf.%s, diags)",
		"utils.FromStringListToTfStringList(ctx, item.%s, &diags)",
	},
	"[]string -> types.Set": {
		"utils.FromStringListPointerToTfStringSet(ctx, &http.%s, diags)",
		"utils.FromTfStringSetToStringList(ctx, tf.%s, diags)",
		"utils.FromStringListPointerToTfStringSet(ctx, &item.%s, &diags)",
	},
	"*[]int -> types.List": {
		"utils.FromIntListPointerToTfInt64List(ctx, http.%s, diags)",
		"utils.TFInt64ListToIntListPointer(ctx, tf.%s, diags)",
		"utils.FromIntListPointerToTfInt64List(ctx, item.%s, &diags)",
	},
	"[]int -> types.List": {
		"utils.FromIntListToTfInt64List(ctx, http.%s, diags)",
		"utils.TFInt64ListToIntList(ctx, tf.%s, diags)",
		"utils.FromIntListToTfInt64List(ctx, item.%s, &diags)",
	},
}

func resolve(m *FieldMapping) {
	pair := m.SDKType + " -> " + m.TFType

	// This pair only populates serialize/deserialize expressions.
	if pair == "[]int -> types.Set" {
		m.SerializeExpr = fmt.Sprintf("utils.FromIntListPointerToTfInt64List(ctx, &http.%s, diags)", m.SDKFieldName)
		m.DeserializeExpr = fmt.Sprintf("utils.TFInt64ListToIntList(ctx, tf.%s, diags)", m.TFFieldName)
		return
	}

	if t, ok := typePairTemplates[pair]; ok {
		applyTemplates(m, t)
		return
	}

	m.IsTODO = true
	m.TODOReason = fmt.Sprintf("unsupported type pair: %s -> %s", m.SDKType, m.TFType)
}

var datasourceParamTemplates = map[string]string{
	"types.String -> *string":               "tf.%s.ValueStringPointer()",
	"types.String -> string":                "tf.%s.ValueString()",
	"types.Bool -> *bool":                   "tf.%s.ValueBoolPointer()",
	"types.Bool -> bool":                    "tf.%s.ValueBool()",
	"types.Int64 -> *int":                   "utils.FromTfInt64ToIntPtr(tf.%s)",
	"types.Int64 -> int":                    "utils.FromTfInt64ToInt(tf.%s)",
	"types.List -> *[]string":               "utils.ConvertTfListToArrayOfString(ctx, tf.%s, diags)",
	"types.List -> []string":                "utils.TfStringListToStringList(ctx, tf.%s, diags)",
	"types.Set -> []string":                 "utils.FromTfStringSetToStringList(ctx, tf.%s, diags)",
	"types.List -> *[]int":                  "utils.ConvertTfListToArrayOfInt(ctx, tf.%s, diags)",
	"types.List -> []int":                   "utils.TFInt64ListToIntList(ctx, tf.%s, diags)",
	"types.List -> *[]AvailabilityZoneName": "utils.ConvertTfListToArrayOfAzName(ctx, tf.%s, diags)",
	"types.List -> *[]ReadVmsParamsTypes":   "utils.ConvertTfListToArrayOfStringEnum[api.ReadVmsParamsTypes](ctx, tf.%s, diags)",
	"PageValue -> *PageParameter":           "func() *api.PageParameter { if utils.IsTfValueNull(tf.%s) { return nil }; return &api.PageParameter{NextToken: tf.%s.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.%s.Size)} }()",
}

func applyDatasourceTemplate(m *FieldMapping, template string) {
	count := strings.Count(template, "%s")
	args := make([]any, count)
	for i := range args {
		args[i] = m.TFFieldName
	}
	m.DatasourceExpr = fmt.Sprintf(template, args...)
}

func resolveDatasourceParams(m *FieldMapping) {
	pair := m.TFType + " -> " + m.SDKType
	if template, ok := datasourceParamTemplates[pair]; ok {
		applyDatasourceTemplate(m, template)
		return
	}
	m.IsTODO = true
	m.TODOReason = fmt.Sprintf("unsupported datasource param pair: %s -> %s", m.TFType, m.SDKType)
}

// ---------- Code generation ----------

func GenerateCreateResponseSerializeBody(fields []FieldMapping) string {
	var parts []string
	for i := range fields {
		f := &fields[i]
		if f.IsTODO {
			parts = append(parts, fmt.Sprintf("\t\t// TODO: auto-mapping failed for %s (%s)", f.TFFieldName, f.TODOReason))
		} else {
			parts = append(parts, fmt.Sprintf("\t\t%s: %s,", f.TFFieldName, f.SerializeExpr))
		}
	}
	return strings.Join(parts, "\n")
}

func GenerateSerializeBody(fields []FieldMapping) string {
	var parts []string
	for i := range fields {
		f := &fields[i]
		if f.IsTODO {
			parts = append(parts, fmt.Sprintf("\t\t// TODO: auto-mapping failed for %s (%s)", f.TFFieldName, f.TODOReason))
		} else {
			parts = append(parts, fmt.Sprintf("\t\t%s: %s,", f.TFFieldName, f.SerializeExpr))
		}
	}
	return strings.Join(parts, "\n")
}

func GenerateDeserializeBody(fields []FieldMapping) string {
	var parts []string
	for i := range fields {
		f := &fields[i]
		if f.TFFieldName == "Id" {
			continue
		}
		if f.NullOnRead {
			// Absent from the read SDK struct; nothing to deserialize.
			continue
		}
		if f.IsTODO {
			parts = append(parts, fmt.Sprintf("\t\t// TODO: auto-mapping failed for %s (%s)", f.TFFieldName, f.TODOReason))
		} else {
			parts = append(parts, fmt.Sprintf("\t\t%s: %s,", f.SDKFieldName, f.DeserializeExpr))
		}
	}
	return strings.Join(parts, "\n")
}

func GenerateDatasourceBody(fields []FieldMapping, hasTags bool) string {
	var parts []string
	for i := range fields {
		f := &fields[i]
		if f.IsTODO {
			parts = append(parts,
				fmt.Sprintf("\t\t// TODO: auto-mapping failed for %s (%s)", f.TFFieldName, f.TODOReason),
				fmt.Sprintf("\t\t// %q: ...,", f.TFSdkTag))
		} else {
			parts = append(parts, fmt.Sprintf("\t\t%q: %s,", f.TFSdkTag, f.DatasourceExpr))
		}
	}
	return strings.Join(parts, "\n")
}

func GenerateDatasourceParamsBody(fields []FieldMapping) string {
	var parts []string
	for i := range fields {
		f := &fields[i]
		if f.IsTODO {
			parts = append(parts, fmt.Sprintf("\t\t// TODO: auto-mapping failed for %s (%s)", f.SDKFieldName, f.TODOReason))
		} else {
			parts = append(parts, fmt.Sprintf("\t\t%s: %s,", f.SDKFieldName, f.DatasourceExpr))
		}
	}
	return strings.Join(parts, "\n")
}

// GenerateNestedSerializeBody creates the attr.Value map body for a sub-serializer.
func GenerateNestedSerializeBody(fields []FieldMapping) string {
	var parts []string
	for i := range fields {
		f := &fields[i]
		if f.IsTODO {
			parts = append(parts,
				fmt.Sprintf("\t\t// TODO: auto-mapping failed for %s (%s)", f.TFFieldName, f.TODOReason),
				fmt.Sprintf("\t\t// %q: ...,", f.TFSdkTag))
		} else {
			parts = append(parts, fmt.Sprintf("\t\t%q: %s,", f.TFSdkTag, f.SerializeExpr))
		}
	}
	return strings.Join(parts, "\n")
}

// GenerateNestedDeserializeBody creates the SDK struct field assignment body for a
// nested deserializer. Unlike GenerateDeserializeBody this does not skip Id.
func GenerateNestedDeserializeBody(fields []FieldMapping) string {
	var parts []string
	for i := range fields {
		f := &fields[i]
		if f.NullOnRead {
			// Absent from the read SDK struct; nothing to deserialize.
			continue
		}
		if f.IsTODO {
			parts = append(parts, fmt.Sprintf("\t\t// TODO: auto-mapping failed for %s (%s)", f.TFFieldName, f.TODOReason))
		} else {
			parts = append(parts, fmt.Sprintf("\t\t%s: %s,", f.SDKFieldName, f.DeserializeExpr))
		}
	}
	return strings.Join(parts, "\n")
}

// dedupeNestedSerializers keeps the first occurrence of each generated function name.
func dedupeNestedSerializers(nested []NestedSerializer) []NestedSerializer {
	seen := make(map[string]bool)
	var out []NestedSerializer
	for i := range nested {
		ns := &nested[i]
		key := ns.SerializeFuncName
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, *ns)
	}
	return out
}

// ---------- Parsing ----------

func collectTypeSpecs(f *ast.File) map[string]*ast.TypeSpec {
	typeSpecs := make(map[string]*ast.TypeSpec)
	ast.Inspect(f, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			typeSpecs[ts.Name.Name] = ts
		}
		return true
	})
	return typeSpecs
}

// resolveSDKStructName returns the real (correctly-cased) SDK type name for a
// requested name, resolving case/separator-insensitive matches via
// resolveSDKTypeName. Used where the generated code must reference the actual
// type name rather than the descriptor's casing convention.
func resolveSDKStructName(path, name string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse SDK file: %w", err)
	}
	typeSpecs := collectTypeSpecs(f)
	return resolveSDKTypeName(typeSpecs, name, path)
}

func parseSDKStruct(path, name string) ([]sdkField, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse SDK file: %w", err)
	}

	typeSpecs := collectTypeSpecs(f)
	resolvedName, err := resolveSDKTypeName(typeSpecs, name, path)
	if err != nil {
		return nil, err
	}

	st, err := resolveStructType(typeSpecs, typeSpecs[resolvedName], resolvedName, path)
	if err != nil {
		return nil, err
	}
	return parseStructFields(st), nil
}

func resolveStructType(typeSpecs map[string]*ast.TypeSpec, ts *ast.TypeSpec, name, path string) (*ast.StructType, error) {
	current := ts
	seen := map[string]bool{name: true}
	for {
		if st, ok := current.Type.(*ast.StructType); ok {
			return st, nil
		}

		ident, ok := current.Type.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("%w %q in %s", errSDKTypeNotStructOrAlias, name, path)
		}
		if seen[ident.Name] {
			return nil, fmt.Errorf("%w %q in %s", errSDKTypeCircularAlias, name, path)
		}
		seen[ident.Name] = true
		next, ok := typeSpecs[ident.Name]
		if !ok {
			return nil, fmt.Errorf("%w %q for %q in %s", errSDKAliasNotFound, ident.Name, name, path)
		}
		current = next
	}
}

func parseStructFields(st *ast.StructType) []sdkField {
	var fields []sdkField
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		fName := field.Names[0].Name
		fType := typeString(field.Type)
		jsonTag := ""
		if field.Tag != nil {
			jsonTag = extractTag(field.Tag.Value, "json")
		}
		fields = append(fields, sdkField{Name: fName, Type: fType, JSONTag: jsonTag})
	}
	return fields
}

// resolveSDKTypeName maps a requested type name to a real SDK type name.
// It prefers an exact match (preserving existing behavior with zero ambiguity
// risk) and only falls back to a case/separator-insensitive match when the
// exact name is absent. The fallback lets descriptors use their own casing
// convention (e.g. "VMs", "VpnConnection", "KubernetesNodepool") without having
// to mirror the vendor SDK's inconsistent casing. An ambiguous normalized match
// is a hard error: the caller must disambiguate explicitly rather than have the
// scaffold guess.
func resolveSDKTypeName(typeSpecs map[string]*ast.TypeSpec, name, path string) (string, error) {
	if _, ok := typeSpecs[name]; ok {
		return name, nil
	}

	key := normalizeSDKName(name)
	var matches []string
	for candidate := range typeSpecs {
		if normalizeSDKName(candidate) == key {
			matches = append(matches, candidate)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("%w %q in %s", errSDKStructNotFound, name, path)
	case 1:
		return matches[0], nil
	default:
		sort.Strings(matches)
		return "", fmt.Errorf(
			"%w %q in %s: matches %v (case/separator-insensitive); set sdk_singular/sdk_plural/params_type explicitly to disambiguate",
			errSDKTypeAmbiguous,
			name,
			path,
			matches,
		)
	}
}

// normalizeSDKName lowercases and strips underscores so that names differing
// only by casing or separators collapse to the same key (digits are preserved).
func normalizeSDKName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

func parseTFModel(path, name string) ([]tfField, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse TF model file: %w", err)
	}

	var fields []tfField
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok || ts.Name.Name != name {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		found = true
		fields = extractTFModelFields(st)
		return false
	})

	if !found {
		return nil, fmt.Errorf("%w %q in %s", errTFModelNotFound, name, path)
	}
	return fields, nil
}

func extractTFModelFields(st *ast.StructType) []tfField {
	var fields []tfField
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		fName := field.Names[0].Name
		// Skip unexported fields and fields without a tfsdk tag.
		if !ast.IsExported(fName) {
			continue
		}
		tfSdkTag := ""
		if field.Tag != nil {
			tfSdkTag = extractTag(field.Tag.Value, "tfsdk")
		}
		if tfSdkTag == "" {
			continue
		}
		fType := normalizeTFType(typeString(field.Type))
		fields = append(fields, tfField{Name: fName, Type: fType, TFSdkTag: tfSdkTag})
	}
	return fields
}

func normalizeTFType(t string) string {
	switch t {
	case "basetypes.StringValue":
		return "types.String"
	case "basetypes.BoolValue":
		return "types.Bool"
	case "basetypes.Int64Value":
		return "types.Int64"
	case "basetypes.Float64Value":
		return "types.Float64"
	case "basetypes.NumberValue":
		return "types.Number"
	case "basetypes.ListValue":
		return "types.List"
	case "basetypes.SetValue":
		return "types.Set"
	case "basetypes.MapValue":
		return "types.Map"
	case "basetypes.ObjectValue":
		return tfTypeObject
	}
	return t
}

// ---------- Helpers ----------

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	default:
		return ""
	}
}

func extractTag(raw, key string) string {
	tag := strings.Trim(raw, "`")
	for _, part := range strings.Split(tag, " ") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, key+":\"") {
			part = strings.TrimPrefix(part, key+":\"")
			if idx := strings.Index(part, "\""); idx != -1 {
				return part[:idx]
			}
		}
	}
	return ""
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		parts[i] = titleCaser.String(p)
	}
	return strings.Join(parts, "")
}

// HasNonTODOMappings returns true if at least one mapping is fully resolved.
func HasNonTODOMappings(fields []FieldMapping) bool {
	for i := range fields {
		if !fields[i].IsTODO {
			return true
		}
	}
	return false
}

// SortedByTFName returns a copy sorted by TF field name for deterministic output.
func SortedByTFName(fields []FieldMapping) []FieldMapping {
	cp := append([]FieldMapping(nil), fields...)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].TFFieldName < cp[j].TFFieldName
	})
	return cp
}

// MappingNeedsCtxDiags reports whether any generated expression selected by expr
// requires a context and diagnostics pointer.
func MappingNeedsCtxDiags(fields []FieldMapping, expr func(FieldMapping) string) bool {
	for i := range fields {
		e := expr(fields[i])
		if strings.Contains(e, "ctx,") || strings.Contains(e, "diags") {
			return true
		}
	}
	return false
}

// nestedDeserializeBodyNeedsCtxDiags reports whether any field expression in a
// nested object deserialize body references ctx or diags.
func nestedDeserializeBodyNeedsCtxDiags(fields []FieldMapping) (needsCtx, needsDiags bool) {
	for i := range fields {
		if strings.Contains(fields[i].DeserializeExpr, "ctx,") {
			needsCtx = true
		}
		if strings.Contains(fields[i].DeserializeExpr, "diags") {
			needsDiags = true
		}
	}
	return
}

// nestedDeserializeBodyNeedsDiags reports whether any field expression in a
// nested set/list deserialize body references diags.
func nestedDeserializeBodyNeedsDiags(fields []FieldMapping) bool {
	for i := range fields {
		if strings.Contains(fields[i].DeserializeExpr, "diags") {
			return true
		}
	}
	return false
}

// buildNestedDeserializeCall builds a helper call with ctx and/or diags omitted
// when the helper does not need them, keeping generated call sites in sync with
// the generated helper signature.
func buildNestedDeserializeCall(funcName, tfFieldName string, needsCtx, needsDiags bool) string {
	var args []string
	if needsCtx {
		args = append(args, "ctx")
	}
	args = append(args, "tf."+tfFieldName)
	if needsDiags {
		args = append(args, "diags")
	}
	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
}

// MappingNeedsDiags reports whether any datasource item expression requires a local diag.Diagnostics variable.
func MappingNeedsDiags(fields []FieldMapping) bool {
	for i := range fields {
		if strings.Contains(fields[i].DatasourceExpr, "diags") {
			return true
		}
	}
	return false
}

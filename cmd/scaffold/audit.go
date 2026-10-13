package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"terraform-provider-numspot/internal/cli/ui"
)

// errAuditUnresolvedMappings is returned when the scaffold audit finds
// descriptors with unresolved mapping overrides.
var errAuditUnresolvedMappings = errors.New("audit failed: one or more scaffold_check descriptors have unresolved mappings")

type auditResult struct {
	DescriptorPath string
	Warnings       []string
	Errors         []string
}

func (r *auditResult) add(format string, args ...interface{}) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

func (r *auditResult) addError(format string, args ...interface{}) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func descriptorPaths(projectDir, descriptorPath string) ([]string, error) {
	if descriptorPath != "" {
		if !filepath.IsAbs(descriptorPath) {
			descriptorPath = filepath.Join(projectDir, descriptorPath)
		}
		return []string{descriptorPath}, nil
	}
	matches, err := filepath.Glob(filepath.Join(projectDir, "internal", "services", "*", "descriptor.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob descriptor files: %w", err)
	}
	sort.Strings(matches)
	return matches, nil
}

func auditPath(projectDir, p string) (auditResult, error) {
	d, err := ParseDescriptor(p)
	if err != nil {
		return auditResult{}, fmt.Errorf("parse %s: %w", p, err)
	}
	if err := d.LoadCoreSignatures(projectDir); err != nil {
		return auditResult{}, fmt.Errorf("load core signatures for %s: %w", p, err)
	}
	return auditDescriptor(d, projectDir, p), nil
}

type auditAccumulator struct {
	results   []auditResult
	hasErrors bool
	clean     int
	warnings  int
	errors    int
}

func (a *auditAccumulator) add(res auditResult) {
	if len(res.Errors) > 0 {
		a.hasErrors = true
	}
	a.errors += len(res.Errors)
	a.warnings += len(res.Warnings)
	if len(res.Warnings) == 0 && len(res.Errors) == 0 {
		a.clean++
		return
	}
	a.results = append(a.results, res)
}

func runAudit(projectDir, descriptorPath string) error {
	paths, err := descriptorPaths(projectDir, descriptorPath)
	if err != nil {
		return err
	}

	ui.Header("Auditing descriptors")

	var acc auditAccumulator
	for i, p := range paths {
		rel, _ := filepath.Rel(projectDir, p)
		ui.Step(i+1, len(paths), "%s", rel)

		res, err := auditPath(projectDir, p)
		if err != nil {
			return err
		}
		acc.add(res)
	}

	for _, res := range acc.results {
		rel, _ := filepath.Rel(projectDir, res.DescriptorPath)
		ui.Infof("%s", rel)
		for _, e := range res.Errors {
			ui.Errorf("%s", e)
		}
		for _, w := range res.Warnings {
			ui.Warnf("%s", w)
		}
	}

	ui.Summary(acc.clean, acc.warnings, acc.errors)

	if acc.hasErrors {
		ui.Errorf("%v", errAuditUnresolvedMappings)
		return errAuditUnresolvedMappings
	}
	ui.Successf("audit passed")
	return nil
}

func auditDescriptor(d *Descriptor, projectDir, descriptorPath string) auditResult {
	res := auditResult{DescriptorPath: descriptorPath}
	auditScaffoldCheck(d, &res)
	auditOverrides(d, &res)
	auditNaming(d, projectDir, &res)
	auditCoreSignatures(d, &res)
	auditResidualTODOs(d, projectDir, &res)
	auditPackageConvention(d, descriptorPath, &res)
	auditSpecNaming(descriptorPath, &res)
	return res
}

// auditSpecNaming enforces that every *_spec.json / *_spec.yaml in a service
// folder is named <folder>_spec.<ext>. The generate-spec / generate-framework
// targets derive the spec path from RESOURCE_NAME (== the folder), so an
// irregular prefix silently excludes the file from regeneration (and from the
// cmd/inject-schema-overrides post-processor). Treated as a hard error so CI
// catches hand-written drift in the naming.
func auditSpecNaming(descriptorPath string, res *auditResult) {
	dir := filepath.Dir(descriptorPath)
	folder := filepath.Base(dir)
	for _, ext := range []string{"json", "yaml"} {
		matches, err := filepath.Glob(filepath.Join(dir, "*_spec."+ext))
		if err != nil {
			continue
		}
		want := folder + "_spec." + ext
		for _, m := range matches {
			if got := filepath.Base(m); got != want {
				res.addError(
					"spec file %q must be named %q (prefix must equal the service folder so generate-spec/generate-framework and inject-schema-overrides pick it up)",
					got,
					want,
				)
			}
		}
	}
}

// auditPackageConvention enforces that the Go package folder is the concatenated
// lowercase form of name (no underscores) and that no redundant package override
// remains. This keeps service package names Go-idiomatic and derivable.
func auditPackageConvention(d *Descriptor, descriptorPath string, res *auditResult) {
	derived := strings.ReplaceAll(d.Name, "_", "")
	folder := filepath.Base(filepath.Dir(descriptorPath))

	if d.Package != "" {
		res.add("package override %q is redundant; the folder is derived from name (remove the package field)", d.Package)
	}
	if folder != derived {
		res.addError("package folder %q must be %q (concatenated lowercase of name %q, no underscores)", folder, derived, d.Name)
	}
}

// auditResidualTODOs enforces the invariant that a descriptor opted into the
// scaffold drift guard (scaffold_check: true) must auto-map every field. If
// generation fails or any field is left as a TODO, the resource would silently
// ship with dropped behavior while drift still passes; that is treated as a
// hard error so the audit (and CI) fails.
func auditResidualTODOs(d *Descriptor, projectDir string, res *auditResult) {
	if !d.ScaffoldCheck {
		return
	}

	mappings, err := BuildMappings(d, projectDir)
	if err != nil {
		res.addError("scaffold_check is true but generation fails: %v", err)
		return
	}

	check := func(kind string, fields []FieldMapping) {
		for i := range fields {
			f := &fields[i]
			// The generator never emits Id in the create request body
			// (GenerateDeserializeBody skips it), so a TODO there is not real.
			if kind == mappingKindDeserialize && f.TFFieldName == "Id" {
				continue
			}
			if f.IsTODO {
				res.addError(
					"%s mapping for %q is unresolved (%s); set scaffold_check: false with a reason, or add a mapping override",
					kind,
					f.TFFieldName,
					f.TODOReason,
				)
			}
		}
	}

	check(mappingKindSerialize, mappings.SerializeFields)
	check(mappingKindCreateResponse, mappings.CreateResponseFields)
	check(mappingKindDeserialize, mappings.DeserializeFields)
	check(mappingKindDatasource, mappings.DatasourceFields)
	check(mappingKindDatasourceParams, mappings.DatasourceParams)
	for i := range mappings.NestedSerializers {
		ns := &mappings.NestedSerializers[i]
		check("nested "+ns.TFFieldName, ns.Fields)
	}
	for i := range mappings.NestedDatasourceSerializers {
		ns := &mappings.NestedDatasourceSerializers[i]
		check("datasource nested "+ns.TFFieldName, ns.Fields)
	}
}

func auditScaffoldCheck(d *Descriptor, res *auditResult) {
	if !d.ScaffoldCheck {
		if strings.TrimSpace(d.Reason) == "" {
			res.add("scaffold_check is disabled but no reason is provided")
		}
		return
	}
	if strings.TrimSpace(d.Reason) != "" {
		res.add("reason is only meaningful when scaffold_check is disabled")
	}
}

func auditOverrides(d *Descriptor, res *auditResult) {
	kindFields := map[string][]OverrideConfig{
		mappingKindSerialize:      d.MappingOverrides.Serialize,
		mappingKindCreateResponse: d.MappingOverrides.CreateResponse,
		mappingKindDeserialize:    d.MappingOverrides.Deserialize,
		mappingKindDatasource:     d.MappingOverrides.Datasource,
	}
	for kind, overrides := range kindFields {
		for i := range overrides {
			auditOverride(d, res, kind, &overrides[i])
		}
	}
	if d.IDSource == "" {
		for _, o := range d.MappingOverrides.Serialize {
			if strings.EqualFold(o.TFField, "Id") && o.SDKField != "" {
				res.add("consider id_source: %s instead of per-kind Id overrides", o.SDKField)
			}
		}
	}
}

func auditOverride(d *Descriptor, res *auditResult, kind string, o *OverrideConfig) {
	if strings.TrimSpace(o.Reason) == "" {
		res.add("mapping_overrides.%s[%s] is missing a reason", kind, o.TFField)
	}
	switch {
	case strings.EqualFold(o.TFField, "Id") && o.SDKField != "" && d.IDSource == "" && kind != mappingKindDeserialize:
		res.add("mapping_overrides.%s[%s] maps Id to %s; use id_source instead", kind, o.TFField, o.SDKField)
	case o.Skip && kind == mappingKindDeserialize:
		auditDeserializeSkip(d, res, o)
	}
}

func auditDeserializeSkip(d *Descriptor, res *auditResult, o *OverrideConfig) {
	switch {
	case d.IsComputedField(o.TFField):
		res.add("mapping_overrides.deserialize[%s] skip is redundant with computed_fields", o.TFField)
	case d.IsExcludeFromRequest(o.TFField):
		res.add("mapping_overrides.deserialize[%s] skip is redundant with exclude_from_request", o.TFField)
	case d.IsManagedByOtherResource(o.TFField):
		res.add("mapping_overrides.deserialize[%s] skip is redundant with managed_by_other_resource", o.TFField)
	}
}

func auditNaming(d *Descriptor, projectDir string, res *auditResult) {
	expectedResourceConstructor := d.ResourceConstructor()
	expectedDatasourceConstructor := d.DatasourceConstructor()
	snake := singularSnake(d.Singular)
	resFile := filepath.Join(projectDir, "internal", "services", d.PackageName(), "resource_"+snake+".go")
	if _, err := os.Stat(resFile); err == nil {
		if actual := findFuncName(resFile, "resource"); actual != "" && actual != expectedResourceConstructor {
			res.add("resource constructor %s should be %s", actual, expectedResourceConstructor)
		}
	}
	dsFile := filepath.Join(projectDir, "internal", "services", d.PackageName(), "datasource_"+snake+"s.go")
	if _, err := os.Stat(dsFile); err == nil {
		if actual := findFuncName(dsFile, "datasource"); actual != "" && actual != expectedDatasourceConstructor {
			res.add("datasource constructor %s should be %s", actual, expectedDatasourceConstructor)
		}
	}
}

func auditCoreSignatures(d *Descriptor, res *auditResult) {
	if d.inferredCore == nil {
		return
	}
	prefix := d.CoreFunctionPrefix()
	expectedRead := "Read" + prefix
	if d.Core.Read.FunctionName != "" {
		expectedRead = d.Core.Read.FunctionName
	}
	suffix := d.Core.ReadList.FunctionSuffix
	expectedReadList := "Read" + d.Plural + suffix
	if d.Core.ReadList.FunctionName != "" {
		expectedReadList = d.Core.ReadList.FunctionName
	}

	auditName(res, "create", d.inferredCore.CreateFuncName, "Create"+prefix)
	auditName(res, "read", d.inferredCore.ReadFuncName, expectedRead)
	auditName(res, "delete", d.inferredCore.DeleteFuncName, "Delete"+prefix)
	auditName(res, "update-tags", d.inferredCore.UpdateTagsFuncName, "Update"+prefix+"Tags")
	auditName(res, "read-list", d.inferredCore.ReadListFuncName, expectedReadList)
}

func auditName(res *auditResult, kind, actual, expected string) {
	if actual != "" && actual != expected {
		res.add("core %s function %s should be %s", kind, actual, expected)
	}
}

func findFuncName(path, kind string) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return ""
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name == nil {
			continue
		}
		if nameMatchesKind(fn.Name.Name, kind) {
			return fn.Name.Name
		}
	}
	return ""
}

func nameMatchesKind(name, kind string) bool {
	switch kind {
	case "resource":
		return strings.HasPrefix(name, "New") && strings.HasSuffix(name, "Resource")
	case "datasource":
		return strings.HasPrefix(name, "New") && strings.HasSuffix(name, "sDataSource")
	}
	return false
}

package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

//go:embed templates
var templatesFS embed.FS

type Generator struct {
	Descriptor *Descriptor
	ProjectDir string
	DryRun     bool
}

func NewGenerator(d *Descriptor, projectDir string, dryRun bool) *Generator {
	return &Generator{
		Descriptor: d,
		ProjectDir: projectDir,
		DryRun:     dryRun,
	}
}

type TemplateData struct {
	*Descriptor
	SerializeBody                        string
	CreateResponseSerializeBody          string
	DeserializeBody                      string
	DatasourceBody                       string
	DatasourceParamsBody                 string
	HasReadListParams                    bool
	HasSerializeMappings                 bool
	HasCreateResponseSerializeMappings   bool
	HasDeserializeMappings               bool
	HasDatasourceMappings                bool
	SerializeNeedsCtxDiags               bool
	CreateResponseSerializeNeedsCtxDiags bool
	DeserializeNeedsCtxDiags             bool
	DatasourceNeedsDiags                 bool
	NestedSerializers                    []NestedSerializer
	HasNestedSerializers                 bool
	NestedSerializersNeedBasetypes       bool
	SerializeNestedSerializers           []NestedSerializer
	SerializeNestedSerializersMap        map[string]bool
	DeserializeNestedSerializers         []NestedSerializer
	DeserializeNestedSerializersMap      map[string]bool
	NestedDatasourceSerializers          []NestedSerializer
	HasNestedDatasourceSerializers       bool
	CreateRequestTypeName                string
	// CreatePreserveFromPlan lists preserved fields the create response does
	// not carry, so the Create method copies them from the plan.
	CreatePreserveFromPlan []string
	// ResourceUsesUtils is true when any generated resource body references the
	// utils package, so the import is included even without tags/ctx-diags.
	ResourceUsesUtils bool
	// CreateNoRequestBody is true when the create API takes no JSON body; the
	// Create call then omits the deserializer and no deserialize func is emitted.
	CreateNoRequestBody bool
	// UpdateCalls are non-tag attribute update calls emitted in the generated
	// Update, each gated on its when_changed fields.
	UpdateCalls []UpdateCallData
	// HasAttributeUpdate is true when UpdateCalls is non-empty.
	HasAttributeUpdate bool
}

// UpdateCallData is the template view of a core.update.calls entry: the core
// function to call, the precomputed plan/state change condition, and the
// comma-joined argument list.
type UpdateCallData struct {
	FunctionName string
	Condition    string
	Args         string
}

func (g *Generator) Generate() error {
	if g.Descriptor.HasResource {
		if g.shouldGenerateCore() {
			if err := g.generateFile("core.go.tmpl", g.corePath()); err != nil {
				return err
			}
		}
		if err := g.generateFile("resource.go.tmpl", g.resourcePath()); err != nil {
			return err
		}
	} else if g.Descriptor.HasDatasource {
		if g.shouldGenerateCore() {
			if err := g.generateFile("core_datasource_only.go.tmpl", g.corePath()); err != nil {
				return err
			}
		}
	}

	if g.Descriptor.HasDatasource {
		if err := g.generateFile("datasource.go.tmpl", g.datasourcePath()); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) shouldGenerateCore() bool {
	if g.Descriptor.Core.Regenerate {
		return true
	}
	// Preserve existing core files unless explicitly regenerating.
	_, err := os.Stat(g.corePath())
	return os.IsNotExist(err)
}

func (g *Generator) generateFile(templateName, outputPath string) error {
	content, err := g.executeTemplate(templateName)
	if err != nil {
		return fmt.Errorf("executing template %s: %w", templateName, err)
	}

	formatted, err := format.Source([]byte(content))
	if err != nil {
		return fmt.Errorf("formatting %s: %w", templateName, err)
	}

	if g.DryRun {
		formatted, err = formatContentWithProjectFormatter(formatted)
		if err != nil {
			return fmt.Errorf("applying project formatter to dry-run %s: %w", outputPath, err)
		}
		fmt.Printf("=== %s ===\n%s\n\n", outputPath, string(formatted))
		return nil
	}

	dir := filepath.Dir(outputPath)
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	if err = os.WriteFile(outputPath, formatted, 0o644); err != nil {
		return fmt.Errorf("writing file %s: %w", outputPath, err)
	}

	if err = formatWithProjectFormatter(outputPath); err != nil {
		return fmt.Errorf("applying project formatter to %s: %w", outputPath, err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

// formatWithProjectFormatter runs golangci-lint fmt on the generated file so
// the output matches the project's gofumpt/gci configuration and does not
// drift after `make fmt`.
func formatWithProjectFormatter(path string) error {
	cmd := exec.CommandContext(context.Background(), "golangci-lint", "fmt", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("golangci-lint fmt %s: %w", path, err)
	}
	return nil
}

// formatContentWithProjectFormatter formats the supplied Go source using the
// project's golangci-lint formatter configuration via stdin. Used for dry-run
// output so it matches the on-disk format used by check-scaffold-drift.
func formatContentWithProjectFormatter(content []byte) ([]byte, error) {
	cmd := exec.CommandContext(context.Background(), "golangci-lint", "fmt", "--stdin")
	cmd.Stdin = bytes.NewReader(content)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("golangci-lint fmt --stdin: %w", err)
	}
	return out.Bytes(), nil
}

func (g *Generator) executeTemplate(name string) (string, error) {
	tmplData, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return "", fmt.Errorf("reading template %s: %w", name, err)
	}

	funcMap := sprig.TxtFuncMap()
	funcMap["GenerateNestedSerializeBody"] = GenerateNestedSerializeBody
	funcMap["GenerateNestedDeserializeBody"] = GenerateNestedDeserializeBody
	funcMap["GenerateDatasourceBody"] = GenerateDatasourceBody

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(tmplData))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", name, err)
	}

	data, err := g.buildTemplateData()
	if err != nil {
		return "", fmt.Errorf("building template data: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %s: %w", name, err)
	}

	return buf.String(), nil
}

func (g *Generator) buildTemplateData() (*TemplateData, error) {
	data := &TemplateData{
		Descriptor: g.Descriptor,
	}

	mappings, err := BuildMappings(g.Descriptor, g.ProjectDir)
	if err != nil {
		return nil, err
	}

	if g.Descriptor.HasResource {
		g.buildResourceData(data, mappings)
	}

	if g.Descriptor.HasDatasource {
		g.buildDatasourceData(data, mappings)
	}

	return data, nil
}

func (g *Generator) buildResourceData(data *TemplateData, mappings *MappingResult) {
	data.SerializeBody = GenerateSerializeBody(mappings.SerializeFields)
	data.DeserializeBody = GenerateDeserializeBody(mappings.DeserializeFields)
	data.HasSerializeMappings = HasNonTODOMappings(mappings.SerializeFields)
	data.HasDeserializeMappings = HasNonTODOMappings(mappings.DeserializeFields)
	data.CreateRequestTypeName = mappings.CreateRequestType
	data.CreateNoRequestBody = g.Descriptor.Core.Create.NoRequestBody
	data.SerializeNeedsCtxDiags = g.Descriptor.HasTags ||
		MappingNeedsCtxDiags(mappings.SerializeFields, func(f FieldMapping) string { return f.SerializeExpr }) ||
		len(mappings.NestedSerializers) > 0
	data.DeserializeNeedsCtxDiags = MappingNeedsCtxDiags(mappings.DeserializeFields, func(f FieldMapping) string { return f.DeserializeExpr })
	data.NestedSerializers = mappings.NestedSerializers
	data.HasNestedSerializers = len(mappings.NestedSerializers) > 0
	data.SerializeNestedSerializers = reachableNestedSerializers(
		mappings.SerializeFields,
		mappings.CreateResponseFields,
		mappings.NestedSerializers,
	)
	data.SerializeNestedSerializersMap = make(map[string]bool, len(data.SerializeNestedSerializers))
	for i := range data.SerializeNestedSerializers {
		ns := &data.SerializeNestedSerializers[i]
		data.SerializeNestedSerializersMap[ns.SerializeFuncName] = true
	}
	data.DeserializeNestedSerializers = reachableNestedSerializers(mappings.DeserializeFields, nil, mappings.NestedSerializers)
	data.DeserializeNestedSerializersMap = make(map[string]bool, len(data.DeserializeNestedSerializers))
	for i := range data.DeserializeNestedSerializers {
		ns := &data.DeserializeNestedSerializers[i]
		data.DeserializeNestedSerializersMap[ns.DeserializeFuncName] = true
	}
	data.NestedSerializersNeedBasetypes = nestedSerializersNeedBasetypes(data.DeserializeNestedSerializers)

	if g.Descriptor.CreateResponseType != "" {
		data.CreateResponseSerializeBody = GenerateCreateResponseSerializeBody(mappings.CreateResponseFields)
		data.HasCreateResponseSerializeMappings = HasNonTODOMappings(mappings.CreateResponseFields)
		data.CreateResponseSerializeNeedsCtxDiags = MappingNeedsCtxDiags(
			mappings.CreateResponseFields,
			func(f FieldMapping) string { return f.SerializeExpr },
		) ||
			len(mappings.NestedSerializers) > 0
	}

	data.CreatePreserveFromPlan = buildCreatePreserveFromPlan(mappings.CreateResponseFields, g.Descriptor.PreserveFromState)
	data.ResourceUsesUtils = resourceUsesUtils(data)
	data.UpdateCalls = buildUpdateCalls(g.Descriptor.Core.Update.Calls)
	data.HasAttributeUpdate = len(data.UpdateCalls) > 0
}

func (g *Generator) buildDatasourceData(data *TemplateData, mappings *MappingResult) {
	data.DatasourceBody = GenerateDatasourceBody(mappings.DatasourceFields, g.Descriptor.HasTags)
	data.DatasourceParamsBody = GenerateDatasourceParamsBody(mappings.DatasourceParams)
	data.HasReadListParams = len(mappings.DatasourceParams) > 0
	data.HasDatasourceMappings = HasNonTODOMappings(mappings.DatasourceFields)
	data.DatasourceNeedsDiags = MappingNeedsDiags(mappings.DatasourceFields) || len(mappings.NestedDatasourceSerializers) > 0
	data.NestedDatasourceSerializers = mappings.NestedDatasourceSerializers
	data.HasNestedDatasourceSerializers = len(mappings.NestedDatasourceSerializers) > 0
}

func buildCreatePreserveFromPlan(createResponseFields []FieldMapping, preserveFromState []string) []string {
	mappedInCreate := make(map[string]bool, len(createResponseFields))
	for i := range createResponseFields {
		m := &createResponseFields[i]
		if !m.IsTODO {
			mappedInCreate[m.TFFieldName] = true
		}
	}

	var preserve []string
	for _, f := range preserveFromState {
		if !mappedInCreate[f] {
			preserve = append(preserve, f)
		}
	}
	return preserve
}

func resourceUsesUtils(data *TemplateData) bool {
	return strings.Contains(data.SerializeBody, "utils.") ||
		strings.Contains(data.DeserializeBody, "utils.") ||
		strings.Contains(data.CreateResponseSerializeBody, "utils.")
}

func buildUpdateCalls(calls []CoreUpdateCall) []UpdateCallData {
	updateCalls := make([]UpdateCallData, 0, len(calls))
	for _, c := range calls {
		conds := make([]string, 0, len(c.WhenChanged))
		for _, f := range c.WhenChanged {
			conds = append(conds, fmt.Sprintf("!plan.%s.Equal(state.%s)", f, f))
		}
		updateCalls = append(updateCalls, UpdateCallData{
			FunctionName: c.FunctionName,
			Condition:    strings.Join(conds, " || "),
			Args:         strings.Join(c.Args, ", "),
		})
	}
	return updateCalls
}

func (g *Generator) corePath() string {
	return findCoreFile(g.ProjectDir, g.Descriptor.Name, g.Descriptor.Singular, g.Descriptor.Plural)
}

func (g *Generator) resourcePath() string {
	snake := singularSnake(g.Descriptor.Singular)
	return filepath.Join(g.ProjectDir, "internal", "services", g.Descriptor.PackageName(), "resource_"+snake+".go")
}

func (g *Generator) datasourcePath() string {
	snake := singularSnake(g.Descriptor.Singular)
	// The datasource file uses the plural form. Most singular snakes need an "s"
	// suffix (snapshot -> snapshots), but a singular already ending in "s"
	// (dhcp_options) must not be doubled.
	if !strings.HasSuffix(snake, "s") {
		snake += "s"
	}
	return filepath.Join(g.ProjectDir, "internal", "services", g.Descriptor.PackageName(), "datasource_"+snake+".go")
}

func nestedSerializersNeedBasetypes(nested []NestedSerializer) bool {
	for i := range nested {
		ns := &nested[i]
		if ns.CollectionKind == collectionKindObject && ns.TFOwnerType == tfTypeObject {
			return true
		}
	}
	return false
}

// reachableNestedSerializers returns the subset of all nested serializers that
// are transitively reachable from the supplied top-level field mappings. The
// same serializer can be reachable from multiple roots (serialize vs create
// response vs deserialize), so callers pass the relevant root field lists.
func reachableNestedSerializers(roots, extraRoots []FieldMapping, all []NestedSerializer) []NestedSerializer {
	byName := make(map[string]NestedSerializer, len(all))
	for i := range all {
		ns := &all[i]
		byName[ns.DeserializeFuncName] = *ns
	}

	seen := make(map[string]bool)
	var walk func(fields []FieldMapping)
	walk = func(fields []FieldMapping) {
		for i := range fields {
			f := &fields[i]
			if f.Nested == nil {
				continue
			}
			name := f.Nested.DeserializeFuncName
			if seen[name] {
				continue
			}
			seen[name] = true
			if ns, ok := byName[name]; ok {
				walk(ns.Fields)
			}
		}
	}
	walk(roots)
	walk(extraRoots)

	var out []NestedSerializer
	for i := range all {
		ns := &all[i]
		if seen[ns.DeserializeFuncName] {
			out = append(out, *ns)
		}
	}
	return out
}

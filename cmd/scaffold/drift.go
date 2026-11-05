package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/numspot/terraform-provider-numspot/internal/cli/ui"
)

var (
	// errScaffoldDriftDetected is returned when generated output differs from
	// the committed files.
	errScaffoldDriftDetected = errors.New("drift detected")
	// errScaffoldDryRunFailed is returned when the scaffold dry-run fails.
	errScaffoldDryRunFailed = errors.New("scaffold dry-run failed")
	// errInjectIgnoresFailed is returned when the inject-ignores step fails.
	errInjectIgnoresFailed = errors.New("inject-ignores failed")
	// errTfplugingenOpenAPIFailed is returned when the openapi codegen step fails.
	errTfplugingenOpenAPIFailed = errors.New("tfplugingen-openapi failed")
	// errInjectSchemaOverridesFailed is returned when the schema-overrides step
	// fails.
	errInjectSchemaOverridesFailed = errors.New("inject-schema-overrides failed")
	// errCommittedSpecMissing is returned when the committed spec.json cannot be
	// read.
	errCommittedSpecMissing = errors.New("committed spec.json missing")
)

// driftDescriptor is the lightweight view of a descriptor.yaml needed by the
// drift subcommand. The full Descriptor in descriptor.go does not parse
// spec_check, so this minimal struct is kept dedicated to the drift checks.
type driftDescriptor struct {
	Name          string `yaml:"name"`
	ScaffoldCheck bool   `yaml:"scaffold_check"`
	SpecCheck     bool   `yaml:"spec_check"`
}

// Pipeline constants mirror the `generate-spec` target in GNUmakefile. The
// openapi codegen CLI is pinned there for reproducibility; keep this in sync.
const (
	apiOASPath     = "internal/sdk/api/public-oas.yaml"
	osOASPath      = "internal/sdk/objectstorage/public-object-storage.yaml"
	codegenOpenAPI = "github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@v0.3.0"
)

type diffOp struct {
	Kind opKind
	Line string
}

type opKind int

const (
	equal opKind = iota
	insert
	deleteOp
)

type checkable struct {
	path string
	desc driftDescriptor
}

// driftResult records which checks drifted for a single descriptor.
type driftResult struct {
	descriptorPath string
	name           string
	specDrift      bool
	scaffoldDrift  bool
}

func collectCheckableDescriptors(root string) ([]checkable, error) {
	paths, err := filepath.Glob(filepath.Join(root, "internal", "services", "*", "descriptor.yaml"))
	if err != nil {
		return nil, fmt.Errorf("glob descriptors: %w", err)
	}

	// Pre-parse so we know how many descriptors will actually be verified and
	// can render a stable [i/total] progress counter.
	var todo []checkable
	for _, p := range paths {
		d, err := parseDriftDescriptor(p)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", rel(root, p), err)
		}
		if d.ScaffoldCheck || d.SpecCheck {
			todo = append(todo, checkable{path: p, desc: d})
		}
	}
	return todo, nil
}

type driftAccumulator struct {
	clean       int
	failedCount int
	failed      bool
}

func (a *driftAccumulator) add(result *driftResult, err error, relPath string) {
	if err != nil {
		a.failed = true
		a.failedCount++
		ui.Errorf("%s: %v", relPath, err)
		return
	}
	if result.scaffoldDrift || result.specDrift {
		a.failed = true
		a.failedCount++
		if result.scaffoldDrift {
			ui.Errorf("%s: drift detected", relPath)
		}
		if result.specDrift {
			ui.Errorf("%s: spec.json drift", relPath)
		}
		return
	}
	a.clean++
}

// runDrift compares committed generated files against `scaffold gen -dry-run`
// output and regenerated *_spec.json. Only descriptors with scaffold_check: true
// (resource/datasource files) and/or spec_check: true (spec.json) are verified.
// It returns a non-nil error when drift is detected so the caller can exit 1.
func runDrift(root string) error {
	todo, err := collectCheckableDescriptors(root)
	if err != nil {
		return err
	}

	if len(todo) == 0 {
		ui.Infof("No descriptors have scaffold_check or spec_check; nothing to verify.")
		return nil
	}

	ui.Header("Checking scaffold + spec drift")

	var acc driftAccumulator
	for i, c := range todo {
		ui.Step(i+1, len(todo), "%s", rel(root, c.path))
		result, err := checkDescriptorDrift(root, c.path, c.desc, false)
		acc.add(result, err, rel(root, c.path))
	}

	ui.Summary(acc.clean, 0, acc.failedCount)

	if acc.failed {
		ui.Errorf("Drift detected. Regenerate the files (make generate-spec / generate-framework / scaffold) and commit, or update the descriptor.")
		return errScaffoldDriftDetected
	}
	ui.Successf("All scaffold-checked and spec-checked descriptors match their generated files.")
	return nil
}

func collectDrifted(root string, todo []checkable) ([]driftResult, error) {
	var drifted []driftResult
	for i, c := range todo {
		ui.Step(i+1, len(todo), "checking %s", rel(root, c.path))
		result, err := checkDescriptorDrift(root, c.path, c.desc, true)
		if err != nil {
			return nil, fmt.Errorf("check %s: %w", rel(root, c.path), err)
		}
		if result.scaffoldDrift || result.specDrift {
			drifted = append(drifted, *result)
		}
	}
	return drifted, nil
}

// runFix detects scaffold/spec drift and regenerates only the artifacts that
// actually drifted for each resource.
func runFix(root string) error {
	todo, err := collectCheckableDescriptors(root)
	if err != nil {
		return err
	}

	if len(todo) == 0 {
		ui.Infof("No descriptors have scaffold_check or spec_check; nothing to fix.")
		return nil
	}

	ui.Header("Fixing scaffold + spec drift")

	drifted, err := collectDrifted(root, todo)
	if err != nil {
		return err
	}
	if len(drifted) == 0 {
		ui.Successf("No scaffold or spec drift detected.")
		return nil
	}

	sort.Slice(drifted, func(i, j int) bool {
		return drifted[i].name < drifted[j].name
	})

	for _, d := range drifted {
		if err := fixDriftedResource(root, d); err != nil {
			return err
		}
	}

	ui.Successf("Fixed %d resources: %s", len(drifted), strings.Join(driftedNames(drifted), ", "))
	return nil
}

func driftedNames(drifted []driftResult) []string {
	names := make([]string, len(drifted))
	for i, d := range drifted {
		names[i] = d.name
	}
	return names
}

func fixDriftedResource(root string, d driftResult) error {
	ui.Infof("Fixing %s (%s)", d.name, formatDriftKinds(d))
	name := filepath.Base(filepath.Dir(d.descriptorPath))

	if d.specDrift {
		if err := runSpecFix(root, name, d.name); err != nil {
			return err
		}
	}
	if d.scaffoldDrift {
		if err := runMake(root, "scaffold", name); err != nil {
			return fmt.Errorf("fixing %s (scaffold): %w", d.name, err)
		}
	}
	return nil
}

func formatDriftKinds(d driftResult) string {
	parts := make([]string, 0, 2)
	if d.specDrift {
		parts = append(parts, "spec")
	}
	if d.scaffoldDrift {
		parts = append(parts, "scaffold")
	}
	return strings.Join(parts, ", ")
}

func runSpecFix(root, name, resourceName string) error {
	specTarget := "generate-spec"
	if resourceName == "bucket" {
		specTarget = "generate-os-spec"
	}
	if err := runMake(root, specTarget, name); err != nil {
		return fmt.Errorf("fixing %s (%s): %w", resourceName, specTarget, err)
	}
	if err := runMake(root, "generate-framework", name); err != nil {
		return fmt.Errorf("fixing %s (generate-framework): %w", resourceName, err)
	}
	return nil
}

// checkDescriptorDrift runs the scaffold and/or spec drift checks for a single
// descriptor. When quiet is true, per-file diffs and residual TODO messages are
// suppressed so the caller can decide what to regenerate.
func checkDescriptorDrift(root, descPath string, d driftDescriptor, quiet bool) (*driftResult, error) {
	result := &driftResult{
		descriptorPath: descPath,
		name:           d.Name,
	}

	if d.ScaffoldCheck {
		drifted, err := checkDescriptor(root, descPath, quiet)
		if err != nil {
			return nil, err
		}
		result.scaffoldDrift = drifted
	}

	if d.SpecCheck {
		drifted, err := checkSpecDrift(root, descPath, d, quiet)
		if err != nil {
			return nil, err
		}
		result.specDrift = drifted
	}

	return result, nil
}

func parseDriftDescriptor(path string) (driftDescriptor, error) {
	var d driftDescriptor
	data, err := os.ReadFile(path)
	if err != nil {
		return d, fmt.Errorf("failed to read descriptor file: %w", err)
	}
	if err := yaml.Unmarshal(data, &d); err != nil {
		return d, fmt.Errorf("failed to unmarshal descriptor YAML: %w", err)
	}
	return d, nil
}

// checkDescriptor regenerates the scaffold files for a descriptor and returns
// true if any generated file differs from the committed version. When quiet is
// true, diffs and residual TODO messages are not printed.
func checkDescriptor(root, descPath string, quiet bool) (bool, error) {
	cmd := exec.CommandContext(context.Background(), "go", "run", "./cmd/scaffold", "gen", "-descriptor", descPath, "-dry-run")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%w:\n%s", errScaffoldDryRunFailed, out)
	}

	files := parseDryRun(string(out))

	drifted := false
	for path, generated := range files {
		// A scaffold_check: true resource must auto-map every field and have a
		// complete generated body. Any residual TODO (mapping failure or template
		// placeholder) means the committed file silently ships an incomplete
		// implementation, so fail even though committed == generated.
		if bytes.Contains(generated, []byte("TODO")) {
			if !quiet {
				ui.Errorf("residual TODO in generated output: %s", rel(root, path))
				ui.Errorf(
					"  a scaffold_check: true resource must be fully generated; add a pattern flag / override, or set scaffold_check: false with a reason",
				)
			}
			drifted = true
		}
		committed, err := os.ReadFile(path)
		if err != nil {
			if !quiet {
				ui.Errorf("generated file missing on disk: %s", rel(root, path))
			}
			drifted = true
			continue
		}
		if !bytes.Equal(generated, committed) {
			if !quiet {
				ui.Errorf("drift detected: %s", rel(root, path))
				fmt.Fprint(os.Stderr, unifiedDiff(rel(root, path), string(committed), string(generated)))
			}
			drifted = true
		}
	}
	return drifted, nil
}

// checkSpecDrift regenerates <name>_spec.json through the same pipeline as
// `make generate-spec` (inject-ignores -> tfplugingen-openapi ->
// inject-schema-overrides) into a temp file and compares it byte-for-byte with
// the committed spec.json. tfplugingen-openapi is not deterministic across
// versions, so the CLI is pinned (see codegenOpenAPI). Only descriptors with
// spec_check: true are verified, so coverage can be ratcheted up as resources
// are made reproducible (mirrors scaffold_check). When quiet is true, the diff
// is not printed.
func checkSpecDrift(root, descPath string, d driftDescriptor, quiet bool) (bool, error) {
	dir := filepath.Dir(descPath)
	base := filepath.Base(dir)
	specYAML := filepath.Join(dir, base+"_spec.yaml")
	specJSON := filepath.Join(dir, base+"_spec.json")

	oas := filepath.Join(root, apiOASPath)
	if d.Name == "bucket" {
		oas = filepath.Join(root, osOASPath)
	}

	tmp, err := os.MkdirTemp("", "specdrift")
	if err != nil {
		return false, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	injected := filepath.Join(tmp, "injected.yaml")
	regen := filepath.Join(tmp, "regen.json")

	if err := runInjectIgnores(root, specYAML, injected); err != nil {
		return false, err
	}
	if err := runTfplugingenOpenAPI(root, injected, regen, oas); err != nil {
		return false, err
	}
	if err := runInjectSchemaOverrides(root, regen, descPath); err != nil {
		return false, err
	}
	return compareSpecFiles(root, specJSON, regen, quiet)
}

func runInjectIgnores(root, specYAML, injected string) error {
	ignCmd := exec.CommandContext(context.Background(), "go", "run", "./cmd/inject-ignores/", specYAML)
	ignCmd.Dir = root
	ignOut, err := ignCmd.Output()
	if err != nil {
		return fmt.Errorf("%w: %v", errInjectIgnoresFailed, err)
	}
	if err := os.WriteFile(injected, ignOut, 0o644); err != nil {
		return fmt.Errorf("failed to write injected spec file: %w", err)
	}
	return nil
}

func runTfplugingenOpenAPI(root, injected, regen, oas string) error {
	genCmd := exec.CommandContext(context.Background(), "go", "run", codegenOpenAPI, "generate",
		"--config="+injected, "--output="+regen, oas)
	genCmd.Dir = root
	if out, err := genCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w:\n%s", errTfplugingenOpenAPIFailed, out)
	}
	return nil
}

func runInjectSchemaOverrides(root, regen, descPath string) error {
	soCmd := exec.CommandContext(context.Background(), "go", "run", "./cmd/inject-schema-overrides/", regen, descPath)
	soCmd.Dir = root
	if out, err := soCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w:\n%s", errInjectSchemaOverridesFailed, out)
	}
	return nil
}

func compareSpecFiles(root, specJSON, regen string, quiet bool) (bool, error) {
	got, err := os.ReadFile(regen)
	if err != nil {
		return false, fmt.Errorf("failed to read regenerated spec.json: %w", err)
	}
	want, err := os.ReadFile(specJSON)
	if err != nil {
		return false, fmt.Errorf("%w: %v", errCommittedSpecMissing, err)
	}
	if !bytes.Equal(got, want) {
		if !quiet {
			r := rel(root, specJSON)
			ui.Errorf("spec.json drift: %s", r)
			fmt.Fprint(os.Stderr, unifiedDiff(r, string(want), string(got)))
		}
		return true, nil
	}
	return false, nil
}

// runMake invokes a Makefile target with RESOURCE_NAME set to the given value.
// Stdout and stderr are streamed so the user sees generation progress.
func runMake(root, target, resourceName string) error {
	cmd := exec.CommandContext(context.Background(), "make", target, "RESOURCE_NAME="+resourceName)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run make %s: %w", target, err)
	}
	return nil
}

// parseDryRun extracts file blocks from scaffold gen -dry-run output.
// Blocks look like:
//
//	=== /absolute/path/to/file.go ===
//	<content>
//
// The provider patch section is skipped.
func parseDryRun(output string) map[string][]byte {
	files := make(map[string][]byte)
	lines := strings.Split(output, "\n")

	var currentPath string
	var currentContent strings.Builder

	flush := func() {
		if currentPath == "" {
			return
		}
		content := currentContent.String()
		// Dry-run separates file blocks with a blank line; strip that separator.
		content = strings.TrimSuffix(content, "\n\n")
		if content == "" {
			files[currentPath] = []byte{}
		} else {
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			files[currentPath] = []byte(content)
		}
		currentPath = ""
		currentContent.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "=== ") && strings.HasSuffix(trimmed, " ===") {
			flush()
			currentPath = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "=== "), " ==="))
			// Skip the provider patch section; it is not a real file.
			if strings.HasPrefix(currentPath, "Provider patch") {
				currentPath = ""
			}
			continue
		}
		if currentPath != "" {
			currentContent.WriteString(line)
			currentContent.WriteByte('\n')
		}
	}
	flush()
	return files
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}

func normalizedLines(text string) []string {
	lines := strings.SplitAfter(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// unifiedDiff returns a minimal unified-style diff of old vs new.
func unifiedDiff(label, oldText, newText string) string {
	oldLines := normalizedLines(oldText)
	newLines := normalizedLines(newText)

	ops := lcsDiff(oldLines, newLines)
	if len(ops) == 0 {
		return ""
	}

	const contextLines = 3
	var b strings.Builder
	fmt.Fprintf(&b, "--- %s (committed)\n", label)
	fmt.Fprintf(&b, "+++ %s (generated)\n", label)

	hunks := collectHunks(ops, contextLines)
	for _, h := range hunks {
		oldLen, newLen := hunkLengths(h.ops)
		fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", h.oldStart, oldLen, h.newStart, newLen)
		for _, op := range h.ops {
			switch op.Kind {
			case equal:
				fmt.Fprintf(&b, " %s", op.Line)
			case insert:
				fmt.Fprintf(&b, "+%s", op.Line)
			case deleteOp:
				fmt.Fprintf(&b, "-%s", op.Line)
			}
		}
	}
	return b.String()
}

func hunkLengths(ops []diffOp) (oldLen, newLen int) {
	for _, op := range ops {
		if op.Kind != insert {
			oldLen++
		}
		if op.Kind != deleteOp {
			newLen++
		}
	}
	return
}

type hunk struct {
	oldStart int
	newStart int
	ops      []diffOp
}

func markChanged(ops []diffOp) []bool {
	changed := make([]bool, len(ops))
	for i, op := range ops {
		if op.Kind != equal {
			changed[i] = true
		}
	}
	return changed
}

func expandHunkBounds(changed []bool, contextLines int) []bool {
	inHunk := make([]bool, len(changed))
	for i := range changed {
		if changed[i] {
			for j := intMax(0, i-contextLines); j <= intMin(len(changed)-1, i+contextLines); j++ {
				inHunk[j] = true
			}
		}
	}
	return inHunk
}

func advanceIndices(op diffOp, oldIdx, newIdx int) (nextOld, nextNew int) {
	switch op.Kind {
	case deleteOp:
		return oldIdx + 1, newIdx
	case insert:
		return oldIdx, newIdx + 1
	default:
		return oldIdx + 1, newIdx + 1
	}
}

func collectHunks(ops []diffOp, contextLines int) []hunk {
	changed := markChanged(ops)
	inHunk := expandHunkBounds(changed, contextLines)

	var hunks []hunk
	var cur []diffOp
	oldIdx, newIdx := 0, 0
	curOldStart, curNewStart := 0, 0
	started := false

	for i, op := range ops {
		beforeOld, beforeNew := oldIdx, newIdx
		oldIdx, newIdx = advanceIndices(op, oldIdx, newIdx)

		if inHunk[i] {
			if !started {
				curOldStart, curNewStart = beforeOld, beforeNew
				started = true
			}
			cur = append(cur, op)
			continue
		}
		if started {
			hunks = append(hunks, hunk{oldStart: curOldStart + 1, newStart: curNewStart + 1, ops: cur})
			cur = nil
			started = false
		}
	}
	if started {
		hunks = append(hunks, hunk{oldStart: curOldStart + 1, newStart: curNewStart + 1, ops: cur})
	}
	return hunks
}

func computeLCS(oldLines, newLines []string) [][]int {
	m, n := len(oldLines), len(newLines)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			switch {
			case oldLines[i] == newLines[j]:
				dp[i][j] = 1 + dp[i+1][j+1]
			case dp[i+1][j] >= dp[i][j+1]:
				dp[i][j] = dp[i+1][j]
			default:
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	return dp
}

func backtrackLCS(dp [][]int, oldLines, newLines []string) []diffOp {
	m, n := len(oldLines), len(newLines)
	ops := make([]diffOp, 0, m+n)
	i, j := 0, 0
	for i < m && j < n {
		switch {
		case oldLines[i] == newLines[j]:
			ops = append(ops, diffOp{Kind: equal, Line: oldLines[i]})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			ops = append(ops, diffOp{Kind: deleteOp, Line: oldLines[i]})
			i++
		default:
			ops = append(ops, diffOp{Kind: insert, Line: newLines[j]})
			j++
		}
	}
	for i < m {
		ops = append(ops, diffOp{Kind: deleteOp, Line: oldLines[i]})
		i++
	}
	for j < n {
		ops = append(ops, diffOp{Kind: insert, Line: newLines[j]})
		j++
	}
	return ops
}

func lcsDiff(oldLines, newLines []string) []diffOp {
	dp := computeLCS(oldLines, newLines)
	return backtrackLCS(dp, oldLines, newLines)
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

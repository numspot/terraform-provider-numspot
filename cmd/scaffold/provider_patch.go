package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type ProviderPatcher struct {
	ProviderPath string
	Descriptor   *Descriptor
	DryRun       bool
}

func NewProviderPatcher(d *Descriptor, projectDir string, dryRun bool) *ProviderPatcher {
	return &ProviderPatcher{
		ProviderPath: fmt.Sprintf("%s/internal/provider/provider.go", projectDir),
		Descriptor:   d,
		DryRun:       dryRun,
	}
}

func (p *ProviderPatcher) Patch() error {
	if p.DryRun {
		p.printDryRun()
		return nil
	}
	return p.patchProvider()
}

func (p *ProviderPatcher) printDryRun() {
	fmt.Printf("=== Provider patch (dry run) ===\n")
	fmt.Printf("  Add import: \"github.com/numspot/terraform-provider-numspot/internal/services/%s\"\n", p.Descriptor.PackageName())
	if p.Descriptor.HasDatasource {
		fmt.Printf("  Add datasource: %s.%s,\n", p.Descriptor.PackageName(), p.Descriptor.DatasourceConstructor())
	}
	if p.Descriptor.HasResource {
		fmt.Printf("  Add resource: %s.%s,\n", p.Descriptor.PackageName(), p.Descriptor.ResourceConstructor())
	}
	fmt.Println()
}

func (p *ProviderPatcher) patchProvider() error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, p.ProviderPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing provider.go: %w", err)
	}

	importPath := "github.com/numspot/terraform-provider-numspot/internal/services/" + p.Descriptor.PackageName()
	importAdded := addImport(node, importPath)
	datasourceAdded, resourceAdded := p.patchRegistrations(node)

	if !importAdded && !datasourceAdded && !resourceAdded {
		fmt.Println("Provider: no changes needed")
		return nil
	}

	if err := writePatchedProvider(p.ProviderPath, fset, node); err != nil {
		return err
	}
	fmt.Println("Patched: internal/provider/provider.go")
	return nil
}

func (p *ProviderPatcher) patchRegistrations(node *ast.File) (datasourceAdded, resourceAdded bool) {
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == "DataSources" && p.Descriptor.HasDatasource {
			datasourceAdded = addRegistration(fn, p.Descriptor.PackageName(), p.Descriptor.DatasourceConstructor())
		}
		if fn.Name.Name == "Resources" && p.Descriptor.HasResource {
			resourceAdded = addRegistration(fn, p.Descriptor.PackageName(), p.Descriptor.ResourceConstructor())
		}
	}
	return
}

func writePatchedProvider(path string, fset *token.FileSet, node *ast.File) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening provider.go for writing: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err = format.Node(f, fset, node); err != nil {
		return fmt.Errorf("formatting provider.go: %w", err)
	}
	return nil
}

func addImport(node *ast.File, importPath string) bool {
	for _, imp := range node.Imports {
		unquoted := strings.Trim(imp.Path.Value, "\"")
		if unquoted == importPath {
			return false
		}
	}

	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("%q", importPath),
		},
	}

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		lastImportIdx := -1
		for i, spec := range genDecl.Specs {
			if _, ok := spec.(*ast.ImportSpec); ok {
				lastImportIdx = i
			}
		}

		if lastImportIdx >= 0 {
			genDecl.Specs = append(
				genDecl.Specs[:lastImportIdx+1],
				append([]ast.Spec{newImport}, genDecl.Specs[lastImportIdx+1:]...)...,
			)
		} else {
			genDecl.Specs = append(genDecl.Specs, newImport)
		}

		return true
	}

	return false
}

func addRegistration(fn *ast.FuncDecl, packageName, constructor string) bool {
	for _, stmt := range fn.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		for _, expr := range retStmt.Results {
			compLit, ok := expr.(*ast.CompositeLit)
			if !ok {
				continue
			}

			for _, elt := range compLit.Elts {
				selExpr, ok := elt.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				ident, ok := selExpr.X.(*ast.Ident)
				if !ok {
					continue
				}

				if ident.Name == packageName && selExpr.Sel.Name == constructor {
					return false
				}
			}

			newEntry := &ast.SelectorExpr{
				X:   ast.NewIdent(packageName),
				Sel: ast.NewIdent(constructor),
			}

			compLit.Elts = append(compLit.Elts, newEntry)
			return true
		}
	}

	return false
}

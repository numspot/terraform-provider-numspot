package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// CoreSignatures holds function names and parameter types discovered from an
// existing internal/core/<name>.go file. The generator uses these values as
// defaults, falling back to descriptor core.* configuration when the parser
// cannot infer something.
type CoreSignatures struct {
	CreateFuncName   string
	CreateParamTypes []string
	CreateExtraArgs  []CoreExtraArg

	ReadFuncName   string
	ReadReturnType string

	ReadListFuncName   string
	ReadListParamsType string

	DeleteFuncName  string
	DeleteExtraArgs []CoreExtraArg

	UpdateTagsFuncName string
}

type CoreExtraArg struct {
	Name string
	Type string
}

// findCoreFile returns the path to an existing internal/core/<name>.go file,
// trying the descriptor name first, then common fallback patterns.
func findCoreFile(projectDir, name, singular, plural string) string {
	candidates := []string{
		filepath.Join(projectDir, "internal", "core", name+".go"),
	}
	if singular != "" {
		snake := singularSnake(singular)
		candidates = append(candidates,
			filepath.Join(projectDir, "internal", "core", snake+".go"),
			filepath.Join(projectDir, "internal", "core", snake+"s.go"),
		)
	}
	if plural != "" {
		candidates = append(candidates,
			filepath.Join(projectDir, "internal", "core", pascalToSnake(plural)+".go"),
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
}

// ParseCoreSignatures parses the existing internal/core/<name>.go file and
// returns discovered function signatures. If the file does not exist, it returns
// an empty result and no error so that callers can fall back to descriptor defaults.
func ParseCoreSignatures(projectDir string, d *Descriptor) (*CoreSignatures, error) {
	corePath := findCoreFile(projectDir, d.Name, d.Singular, d.Plural)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, corePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse core file %s: %w", corePath, err)
	}

	sig := &CoreSignatures{}
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name == nil {
			return true
		}
		tryInferSignature(sig, fn)
		return true
	})

	return sig, nil
}

func tryInferSignature(sig *CoreSignatures, fn *ast.FuncDecl) {
	params, returns := coreFuncSignature(fn)
	name := fn.Name.Name
	_ = tryInferCreate(sig, name, params, returns) ||
		tryInferRead(sig, name, params, returns) ||
		tryInferReadList(sig, name, params, returns) ||
		tryInferDelete(sig, name, params, returns) ||
		tryInferUpdateTags(sig, name, params, returns)
}

func tryInferCreate(sig *CoreSignatures, name string, params, returns []string) bool {
	if sig.CreateFuncName != "" || !isCreateCandidate(name, params, returns) {
		return false
	}
	sig.CreateFuncName = name
	sig.CreateParamTypes = params
	sig.CreateExtraArgs = extractExtraArgs(params, 3)
	return true
}

func tryInferRead(sig *CoreSignatures, name string, params, returns []string) bool {
	if sig.ReadFuncName != "" || !isReadCandidate(name, params, returns) {
		return false
	}
	sig.ReadFuncName = name
	sig.ReadReturnType = firstReturnType(returns)
	return true
}

func tryInferReadList(sig *CoreSignatures, name string, params, returns []string) bool {
	if sig.ReadListFuncName != "" || !isReadListCandidate(name, params, returns) {
		return false
	}
	sig.ReadListFuncName = name
	sig.ReadListParamsType = extractParamsType(params)
	return true
}

func tryInferDelete(sig *CoreSignatures, name string, params, returns []string) bool {
	if sig.DeleteFuncName != "" || !isDeleteCandidate(name, params, returns) {
		return false
	}
	sig.DeleteFuncName = name
	sig.DeleteExtraArgs = extractExtraArgs(params, 3)
	return true
}

func tryInferUpdateTags(sig *CoreSignatures, name string, params, returns []string) bool {
	if sig.UpdateTagsFuncName != "" || !isUpdateTagsCandidate(name, params, returns) {
		return false
	}
	sig.UpdateTagsFuncName = name
	return true
}

func coreFuncSignature(fn *ast.FuncDecl) (params, returns []string) {
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			params = append(params, typeString(p.Type))
		}
	}
	if fn.Type.Results != nil {
		for _, r := range fn.Type.Results.List {
			returns = append(returns, typeString(r.Type))
		}
	}
	return params, returns
}

func isCreateCandidate(name string, params, returns []string) bool {
	if !strings.HasPrefix(name, "Create") || len(params) < 3 {
		return false
	}
	return params[0] == coreContextType &&
		params[1] == coreClientType &&
		(strings.HasPrefix(params[2], "api.") || strings.HasPrefix(params[2], "*api.")) &&
		len(returns) == 2 &&
		strings.HasPrefix(returns[0], "*api.") &&
		returns[1] == coreErrorType
}

func isReadCandidate(name string, params, returns []string) bool {
	if !strings.HasPrefix(name, "Read") || len(params) < 3 || len(returns) != 2 {
		return false
	}
	return params[0] == coreContextType &&
		params[1] == coreClientType &&
		params[2] == coreStringType &&
		strings.HasPrefix(returns[0], "*api.") &&
		returns[1] == coreErrorType
}

func isReadListCandidate(name string, params, returns []string) bool {
	if !strings.HasPrefix(name, "Read") || len(params) < 2 || len(returns) != 2 {
		return false
	}
	if params[0] != coreContextType || params[1] != coreClientType {
		return false
	}
	if !hasTypedParamsArg(params) {
		return false
	}
	return returnsList(params, returns) && returns[1] == coreErrorType
}

func hasTypedParamsArg(params []string) bool {
	if len(params) < 3 {
		return true
	}
	paramsType := params[2]
	return strings.HasPrefix(paramsType, "api.") || strings.HasPrefix(paramsType, "*api.")
}

func returnsList(params, returns []string) bool {
	// For functions with a typed third argument, require an actual slice return
	// to avoid treating singular reads (with e.g. api.ResourceIdentifier) as lists.
	if len(params) >= 3 {
		return strings.HasPrefix(returns[0], "*[]api.") || strings.HasPrefix(returns[0], "[]api.")
	}
	return strings.HasPrefix(returns[0], "*[]api.") || strings.HasPrefix(returns[0], "[]api.") || isListWrapperType(returns[0])
}

func isDeleteCandidate(name string, params, returns []string) bool {
	if !strings.HasPrefix(name, "Delete") || len(params) < 3 {
		return false
	}
	return params[0] == coreContextType &&
		params[1] == coreClientType &&
		params[2] == coreStringType &&
		len(returns) == 1 &&
		returns[0] == coreErrorType
}

func isUpdateTagsCandidate(name string, params, returns []string) bool {
	if !hasUpdateTagsShape(name, params, returns) {
		return false
	}
	return isTagSliceParam(params[3]) &&
		isTagSliceParam(params[4]) &&
		strings.HasPrefix(returns[0], "*api.") &&
		returns[1] == coreErrorType
}

func hasUpdateTagsShape(name string, params, returns []string) bool {
	return strings.HasPrefix(name, "Update") && strings.HasSuffix(name, "Tags") &&
		len(params) >= 5 && len(returns) == 2 &&
		params[0] == coreContextType && params[1] == coreClientType && params[2] == coreStringType
}

func isTagSliceParam(t string) bool {
	return strings.HasPrefix(t, "[]api.")
}

func firstReturnType(returns []string) string {
	if len(returns) == 0 {
		return ""
	}
	return returns[0]
}

func extractExtraArgs(params []string, skip int) []CoreExtraArg {
	var args []CoreExtraArg
	// Names are not preserved by typeString, so derive a placeholder from the type.
	for i, t := range params {
		if i < skip {
			continue
		}
		args = append(args, CoreExtraArg{Name: fmt.Sprintf("arg%d", i), Type: t})
	}
	return args
}

func extractParamsType(params []string) string {
	if len(params) < 3 {
		return ""
	}
	t := params[2]
	// Strip leading pointer and api package qualifier.
	t = strings.TrimPrefix(t, "*")
	t = strings.TrimPrefix(t, "api.")
	return t
}

// isListWrapperType recognises core functions that return a wrapper struct
// containing an Items field (e.g. *api.PostgresqlListClusters200Response).
func isListWrapperType(typeName string) bool {
	return strings.HasPrefix(typeName, "*api.") && strings.Contains(typeName, "List")
}

// LoadCoreSignatures parses the existing core file and stores the discovered
// signatures on the descriptor. It never writes to the core file.
func (d *Descriptor) LoadCoreSignatures(projectDir string) error {
	sig, err := ParseCoreSignatures(projectDir, d)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing core files are fine; the generator will use defaults.
			return nil
		}
		return err
	}
	d.inferredCore = sig
	return nil
}

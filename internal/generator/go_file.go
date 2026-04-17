package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

type goFile struct {
	content       string
	file          *ast.File
	fset          *token.FileSet
	imports       map[string]struct{}
	importAliases map[string]string
}

type methodSignature struct {
	Params  []string
	Results []string
}

type methodSignatureMismatch struct {
	MethodName string
	Expected   methodSignature
	Actual     methodSignature
}

func parseGoFile(path string) (*goFile, error) {
	content, err := util.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	gf := &goFile{
		content:       content,
		file:          file,
		fset:          fset,
		imports:       make(map[string]struct{}),
		importAliases: make(map[string]string),
	}
	for _, imp := range file.Imports {
		pathValue, err := strconv.Unquote(imp.Path.Value)
		if err == nil {
			gf.imports[pathValue] = struct{}{}
			if alias := importAlias(imp, pathValue); alias != "" {
				gf.importAliases[alias] = pathValue
			}
		}
	}

	return gf, nil
}

func (gf *goFile) HasType(typeName string) bool {
	for _, decl := range gf.file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if ok && ts.Name.Name == typeName {
				return true
			}
		}
	}
	return false
}

func (gf *goFile) MethodNamesForReceiver(receiverName string) map[string]struct{} {
	methods := make(map[string]struct{})
	for _, decl := range gf.file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		if receiverTypeName(fn.Recv.List[0].Type) == receiverName {
			methods[fn.Name.Name] = struct{}{}
		}
	}
	return methods
}

func (gf *goFile) MethodSignaturesForReceiver(receiverName string) map[string]methodSignature {
	methods := make(map[string]methodSignature)
	for _, decl := range gf.file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		if receiverTypeName(fn.Recv.List[0].Type) != receiverName {
			continue
		}
		methods[fn.Name.Name] = methodSignature{
			Params:  fieldListTypeStrings(fn.Type.Params, gf.importAliases, gf.fset),
			Results: fieldListTypeStrings(fn.Type.Results, gf.importAliases, gf.fset),
		}
	}
	return methods
}

func receiverTypeName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return receiverTypeName(x.X)
	default:
		return ""
	}
}

func (gf *goFile) EnsureImports(importPaths []string) {
	missing := make([]string, 0, len(importPaths))
	for _, importPath := range importPaths {
		if importPath == "" {
			continue
		}
		if _, ok := gf.imports[importPath]; ok {
			continue
		}
		gf.imports[importPath] = struct{}{}
		missing = append(missing, importPath)
	}
	if len(missing) == 0 {
		return
	}

	sort.Strings(missing)
	importBlock := buildImportLines(missing)
	for _, decl := range gf.file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}

		if gen.Lparen.IsValid() {
			offset := gf.fset.Position(gen.Rparen).Offset
			gf.content = gf.content[:offset] + importBlock + gf.content[offset:]
			return
		}

		offset := gf.fset.Position(gen.End()).Offset
		gf.content = gf.content[:offset] + "\n\nimport (" + importBlock + "\n)" + gf.content[offset:]
		return
	}

	newline := strings.Index(gf.content, "\n")
	if newline == -1 {
		gf.content += "\n\nimport (" + importBlock + "\n)\n"
		return
	}

	gf.content = gf.content[:newline+1] + "\nimport (" + importBlock + "\n)\n" + gf.content[newline+1:]
}

func (gf *goFile) AppendCode(fragment string) (string, error) {
	gf.content = strings.TrimRight(gf.content, "\r\n") + "\n\n" + strings.TrimSpace(fragment) + "\n"

	formatted, err := util.FormatGoFile(gf.content)
	if err != nil {
		return "", err
	}
	gf.content = formatted
	return formatted, nil
}

func buildImportLines(importPaths []string) string {
	var builder strings.Builder
	for _, importPath := range importPaths {
		builder.WriteString("\n\t")
		builder.WriteString(strconv.Quote(importPath))
	}
	return builder.String()
}

func (ms methodSignature) Equal(other methodSignature) bool {
	if len(ms.Params) != len(other.Params) || len(ms.Results) != len(other.Results) {
		return false
	}
	for i := range ms.Params {
		if ms.Params[i] != other.Params[i] {
			return false
		}
	}
	for i := range ms.Results {
		if ms.Results[i] != other.Results[i] {
			return false
		}
	}
	return true
}

func (ms methodSignature) String() string {
	params := "(" + strings.Join(ms.Params, ", ") + ")"
	switch len(ms.Results) {
	case 0:
		return params
	case 1:
		return params + " " + ms.Results[0]
	default:
		return params + " (" + strings.Join(ms.Results, ", ") + ")"
	}
}

func classifyExpectedMethods(methods []types.Method, existing map[string]methodSignature, config types.Config, protoPackage string) ([]types.Method, []methodSignatureMismatch) {
	missing := make([]types.Method, 0, len(methods))
	mismatched := make([]methodSignatureMismatch, 0)
	for _, method := range methods {
		expected := expectedServiceMethodSignature(method, config, protoPackage)
		actual, ok := existing[method.Name]
		if !ok {
			missing = append(missing, method)
			continue
		}
		if actual.Equal(expected) {
			continue
		}
		mismatched = append(mismatched, methodSignatureMismatch{
			MethodName: method.Name,
			Expected:   expected,
			Actual:     actual,
		})
	}
	return missing, mismatched
}

func newMethodSignatureMismatchError(target string, mismatches []methodSignatureMismatch) error {
	details := make([]string, 0, len(mismatches))
	for _, mismatch := range mismatches {
		details = append(details, fmt.Sprintf("%s 期望 %s，实际 %s", mismatch.MethodName, mismatch.Expected, mismatch.Actual))
	}
	sort.Strings(details)
	return fmt.Errorf("%s存在同名但签名不匹配的方法: %s", target, strings.Join(details, "; "))
}

func expectedServiceMethodSignature(method types.Method, config types.Config, protoPackage string) methodSignature {
	return methodSignature{
		Params: []string{
			"context.Context",
			canonicalProtoType(config, protoPackage, method.RequestPkg, method.RequestType),
		},
		Results: []string{
			canonicalProtoType(config, protoPackage, method.ResponsePkg, method.ResponseType),
			"error",
		},
	}
}

func canonicalProtoType(config types.Config, protoPackage, pkg, typeName string) string {
	localPkg := pkg
	if localPkg == "" {
		localPkg = protoPackage
	}
	importPath := protoImportPath(config, protoPackage, pkg)
	if importPath == "" {
		return "*" + localPkg + "." + typeName
	}
	return "*" + importPath + "." + typeName
}

func importAlias(imp *ast.ImportSpec, pathValue string) string {
	if imp.Name != nil {
		switch imp.Name.Name {
		case "_", ".":
			return ""
		default:
			return imp.Name.Name
		}
	}
	slash := strings.LastIndex(pathValue, "/")
	if slash == -1 {
		return pathValue
	}
	return pathValue[slash+1:]
}

func fieldListTypeStrings(fields *ast.FieldList, importAliases map[string]string, fset *token.FileSet) []string {
	if fields == nil {
		return nil
	}

	types := make([]string, 0, len(fields.List))
	for _, field := range fields.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}

		typeName := canonicalTypeExpr(field.Type, importAliases, fset)
		for i := 0; i < count; i++ {
			types = append(types, typeName)
		}
	}
	return types
}

func canonicalTypeExpr(expr ast.Expr, importAliases map[string]string, fset *token.FileSet) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		if ident, ok := x.X.(*ast.Ident); ok {
			if importPath, ok := importAliases[ident.Name]; ok {
				return importPath + "." + x.Sel.Name
			}
		}
		return canonicalTypeExpr(x.X, importAliases, fset) + "." + x.Sel.Name
	case *ast.StarExpr:
		return "*" + canonicalTypeExpr(x.X, importAliases, fset)
	case *ast.ArrayType:
		prefix := "[]"
		if x.Len != nil {
			prefix = "[" + renderNode(x.Len, fset) + "]"
		}
		return prefix + canonicalTypeExpr(x.Elt, importAliases, fset)
	case *ast.Ellipsis:
		return "..." + canonicalTypeExpr(x.Elt, importAliases, fset)
	case *ast.MapType:
		return "map[" + canonicalTypeExpr(x.Key, importAliases, fset) + "]" + canonicalTypeExpr(x.Value, importAliases, fset)
	case *ast.ChanType:
		switch x.Dir {
		case ast.SEND:
			return "chan<- " + canonicalTypeExpr(x.Value, importAliases, fset)
		case ast.RECV:
			return "<-chan " + canonicalTypeExpr(x.Value, importAliases, fset)
		default:
			return "chan " + canonicalTypeExpr(x.Value, importAliases, fset)
		}
	case *ast.ParenExpr:
		return "(" + canonicalTypeExpr(x.X, importAliases, fset) + ")"
	default:
		return renderNode(expr, fset)
	}
}

func renderNode(node interface{}, fset *token.FileSet) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

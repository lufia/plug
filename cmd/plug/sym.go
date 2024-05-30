package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types/typeutil"
)

type Sym struct {
	pkgPath  string
	typeName string
	funcName string
}

func (m *Sym) PkgPath() string {
	return m.pkgPath
}

func (m *Sym) Func() (typeName, funcName string) {
	return m.typeName, m.funcName
}

func (m *Sym) String() string {
	a := make([]string, 0, 3)
	a = append(a, m.pkgPath)
	if m.typeName != "" {
		a = append(a, m.typeName)
	}
	a = append(a, m.funcName)
	return strings.Join(a, ".")
}

func FindPlugSyms(pkgPath string) ([]*Sym, error) {
	var c loader.Config
	c.ImportWithTests(pkgPath)
	p, err := c.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", pkgPath, err)
	}
	var syms []*Sym
	for _, pkg := range p.InitialPackages() {
		for _, f := range pkg.Files {
			m, err := importMap(f)
			if err != nil {
				return nil, err
			}
			m[pkg.Pkg.Name()] = pkg.Pkg.Path()
			for s := range findPlugSyms(pkg, c.Fset, f, m) {
				if slices.ContainsFunc(syms, func(v *Sym) bool { return *v == *s }) {
					continue
				}
				syms = append(syms, s)
			}
		}
	}
	return syms, nil
}

func findPlugSyms(pkg *loader.PackageInfo, fset *token.FileSet, f *ast.File, m map[string]string) <-chan *Sym {
	c := make(chan *Sym)
	w := walker(func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isPlugFunc(&pkg.Info, call) || len(call.Args) != 2 {
			return true
		}
		if verbose {
			ast.Print(fset, call.Args[1])
		}
		pkgName, typeName, funcName := parseExpr(&pkg.Info, call.Args[1])
		c <- &Sym{
			pkgPath:  m[pkgName],
			typeName: typeName,
			funcName: funcName,
		}
		return true
	})
	go func() {
		ast.Walk(w, f)
		close(c)
	}()
	return c
}

// importMap returns a map of name -> importPath.
func importMap(f *ast.File) (map[string]string, error) {
	m := make(map[string]string)
	for _, p := range f.Imports {
		val := p.Path.Value[1 : len(p.Path.Value)-1]
		if p.Name != nil {
			m[p.Name.Name] = val
			continue
		}
		p, err := build.Default.Import(val, ".", 0)
		if err != nil {
			return nil, err
		}
		m[p.Name] = val
	}
	return m, nil
}

func parseExpr(info *types.Info, expr ast.Expr) (pkgName, typeName, funcName string) {
	switch t := expr.(type) {
	case *ast.SelectorExpr: // X.Sel
		pkgName, typeName, funcName = parseExpr(info, t.X)
		switch p := info.ObjectOf(t.Sel).(type) {
		case *types.Func:
			pkgName = p.Pkg().Name()
			funcName = p.Name()
		case *types.TypeName:
			pkgName = p.Pkg().Name()
			typeName = p.Name()
		}
		return
	case *ast.Ident:
		switch p := info.ObjectOf(t).(type) {
		case *types.PkgName:
			pkgName = p.Name()
		case *types.Func:
			pkgName = p.Pkg().Name()
			funcName = p.Name()
		default:
			pkgName = t.Name
		}
		return
	case *ast.CompositeLit: // Type{}
		return parseExpr(info, t.Type)
	case *ast.ParenExpr: // (X)
		return parseExpr(info, t.X)
	case *ast.UnaryExpr: // &X
		return parseExpr(info, t.X)
	case *ast.IndexExpr: // X[Index]
		return parseExpr(info, t.X)
	case *ast.StarExpr: // *X
		return parseExpr(info, t.X)
	case *ast.CallExpr: // Fun()
		return parseExpr(info, t.Fun)
	default:
		panic(t)
	}
}

func isPlugFunc(info *types.Info, call *ast.CallExpr) bool {
	obj := typeutil.Callee(info, call)
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	if obj.Pkg().Path() != "github.com/lufia/plug" || obj.Name() != "Func" {
		return false
	}
	return true
}

type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}

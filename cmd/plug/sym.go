package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
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

func FindPlugSyms(dir string) ([]*Sym, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("ParseDir(%q): %w", dir, err)
	}

	var syms []*Sym
	for _, pkg := range pkgs {
		var (
			config = types.Config{
				Importer: importer.ForCompiler(fset, "source", nil),
			}
			info = types.Info{
				Types:      make(map[ast.Expr]types.TypeAndValue),
				Defs:       make(map[*ast.Ident]types.Object),
				Uses:       make(map[*ast.Ident]types.Object),
				Selections: make(map[*ast.SelectorExpr]*types.Selection),
			}
		)
		_, err := config.Check(dir, fset, MapValues(pkg.Files), &info)
		if err != nil {
			return nil, err
		}
		for _, f := range pkg.Files {
			m, err := importMap(f)
			if err != nil {
				return nil, err
			}
			for s := range findPlugSyms(&info, fset, f, m) {
				syms = append(syms, s)
			}
		}
	}
	return syms, nil
}

func findPlugSyms(info *types.Info, fset *token.FileSet, f *ast.File, m map[string]string) <-chan *Sym {
	c := make(chan *Sym)
	w := walker(func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isPlugSet(call.Fun) || len(call.Args) != 2 {
			return true
		}
		if verbose {
			ast.Print(fset, call.Args[0])
		}
		pkgName, typeName, funcName := parseExpr(info, call.Args[0])
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
			return
		}
		pkgName = t.Name
		return pkgName, typeName, ""
	case *ast.CompositeLit: // Type{}
		return parseExpr(info, t.Type)
	case *ast.ParenExpr: // (X)
		return parseExpr(info, t.X)
	case *ast.UnaryExpr: // &X
		return parseExpr(info, t.X)
	case *ast.StarExpr: // *X
		return parseExpr(info, t.X)
	case *ast.CallExpr: // Fun()
		return parseExpr(info, t.Fun)
	default:
		panic(t)
	}
}

func isPlugSet(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	p, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return p.Name == "plug" && sel.Sel.Name == "Set"
}

type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}

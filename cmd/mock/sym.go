package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
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

func FindMockSyms(dir string) ([]*Sym, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("ParseDir(%q): %w", dir, err)
	}

	var syms []*Sym
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			m, err := importMap(f)
			if err != nil {
				return nil, err
			}
			for s := range findMockSyms(f, m) {
				syms = append(syms, s)
			}
		}
	}
	return syms, nil
}

func findMockSyms(f *ast.File, m map[string]string) <-chan *Sym {
	c := make(chan *Sym)
	w := walker(func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isMockSet(call.Fun) || len(call.Args) != 2 {
			return true
		}
		pkgName, typeName, funcName := parseExpr(call.Args[0])
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

func parseExpr(expr ast.Expr) (pkgName, typeName, funcName string) {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		pkgName, typeName = typeStr(t.X)
		return pkgName, typeName, t.Sel.Name
	case *ast.Ident:
		return ".", "", t.Name
	default:
		return "", "", ""
	}
}

func typeStr(expr ast.Expr) (pkgName, typeName string) {
	for {
		switch p := expr.(type) {
		case *ast.Ident:
			pkgName = p.Name
			return pkgName, typeName
		case *ast.CompositeLit: // T{}
			expr = p.Type
		case *ast.ParenExpr: // ()
			expr = p.X
		case *ast.UnaryExpr: // &
			expr = p.X
		case *ast.SelectorExpr: // a.b
			typeName = p.Sel.Name
			expr = p.X
		default:
			panic(p)
		}
	}
}

func isMockSet(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	p, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return p.Name == "mock" && sel.Sel.Name == "Set"
}

type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}

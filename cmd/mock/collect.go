package main

import (
	"go/ast"
	"log"
)

type Sym struct {
	pkg  *pkgRef
	name string
}

func (m *Sym) PkgPath() string {
	return m.pkg.pkgPath
}

func (m *Sym) File() string {
	return ""
}

func (m *Sym) Func() (typeName, funcName string) {
	return "", m.name
}

func FindMockSetInPkgs(pkgs map[string]*ast.Package) []*Sym {
	var syms []*Sym
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			m := importMap(f)
			c := findMockSet(f)
			for s := range c {
				log.Println(m[s.pkg.pkgPath], s.pkg.typeName, s.pkg.ind, s.name)
				syms = append(syms, s)
			}
		}
	}
	return syms
}

func importMap(f *ast.File) map[string]string {
	m := make(map[string]string)
	for _, p := range f.Imports {
		val := p.Path.Value[1 : len(p.Path.Value)-1]
		var name string
		if p.Name != nil {
			name = p.Name.Name
		} else {
			name = loadPkg(val).Name
		}
		m[name] = val
	}
	return m
}

func findMockSet(f *ast.File) <-chan *Sym {
	c := make(chan *Sym)
	w := walker(func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isMockSet(call.Fun) || len(call.Args) != 2 {
			return true
		}
		pkg, name := parseExpr(call.Args[0])
		c <- &Sym{
			pkg:  pkg,
			name: name,
		}
		return true
	})
	go func() {
		ast.Walk(w, f)
		close(c)
	}()
	return c
}

type pkgRef struct {
	pkgPath  string
	typeName string
	ind      bool
}

func parseExpr(expr ast.Expr) (*pkgRef, string) {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		return ident(t.X), t.Sel.Name
	case *ast.Ident:
		return &pkgRef{pkgPath: "."}, t.Name
	default:
		return nil, ""
	}
}

func (p *pkgRef) String() string {
	s := p.pkgPath
	if p.typeName != "" {
		s += "."
		s += p.typeName
		if p.ind {
			s += "*"
		}
	}
	return s
}

func ident(expr ast.Expr) *pkgRef {
	var pkg pkgRef
	for {
		switch p := expr.(type) {
		case *ast.Ident:
			pkg.pkgPath = p.Name
			return &pkg
		case *ast.CompositeLit: // T{}
			expr = p.Type
		case *ast.ParenExpr: // ()
			expr = p.X
		case *ast.UnaryExpr: // &
			pkg.ind = true
			expr = p.X
		case *ast.SelectorExpr: // a.b
			pkg.typeName = p.Sel.Name
			expr = p.X
		default:
			log.Printf("Type=%[1]T, Value=%[1]v\n", p)
			return nil
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

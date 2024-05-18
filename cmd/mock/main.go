package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"path"
)

func main() {
	log.SetFlags(0)

	target, err := build.Default.Import(".", ".", 0)
	if err != nil {
		log.Fatal(err)
	}
	fset := token.NewFileSet()
	//pkgs, err := parser.ParseDir(fset, target.Dir, nil, parser.ParseComments)
	pkgs, err := parser.ParseDir(fset, target.Dir, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	findMockSetInPkgs(pkgs)
}

func findMockSetInPkgs(pkgs map[string]*ast.Package) {
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			findMockSet(f)
		}
	}
}

func findMockSet(f *ast.File) {
	imports := make(map[string]string)
	for _, i := range f.Imports {
		// TODO: get real package name
		val := i.Path.Value[1 : len(i.Path.Value)-1]
		var name string
		if i.Name != nil {
			name = i.Name.Name
		} else {
			name = path.Base(val)
		}
		log.Println("Imports:", name, val)
		imports[name] = val
	}

	w := walker(func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isMockSet(call.Fun) || len(call.Args) != 2 {
			return true
		}
		pkg, name := parseExpr(call.Args[0])
		log.Println(imports[pkg.pkgPath], pkg.typeName, pkg.ind, name)
		return true
	})
	ast.Walk(w, f)
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

type pkgRef struct {
	pkgPath  string
	typeName string
	ind      bool
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

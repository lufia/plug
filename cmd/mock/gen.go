package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func Generate(sym *Sym) string {
	pkg := loadPkg(sym.PkgPath())
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkg.Dir, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	p := pkgs[pkg.Name]
	fn := findDecl(p, sym)
	if fn == nil {
		log.Fatal("no func")
	}
	fn.Replace(fset)
	return ""
}

type Func struct {
	decl *ast.FuncDecl
	file *ast.File
}

func (fn *Func) Replace(fset *token.FileSet) {
	name := fn.decl.Name.Name
	fn.decl.Name.Name = "_" + name
	astutil.AddImport(fset, fn.file, "github.com/lufia/mock")
	if err := format.Node(os.Stdout, fset, fn.file); err != nil {
		log.Fatal(err)
	}

	fmt.Print("func ")
	if fn.decl.Recv != nil {
	}
	fmt.Print(name)
	fmt.Print("(")
	printTypeList(os.Stdout, fset, fn.decl.Type.Params)
	fmt.Print(") (")
	printTypeList(os.Stdout, fset, fn.decl.Type.Results)
	fmt.Println(") {")
	fmt.Printf("\tf := mock.Get(%s)\n", name)
	fmt.Println("\treturn f(d)")
	fmt.Println("}")
}

func printTypeList(w io.Writer, fset *token.FileSet, l *ast.FieldList) {
	if l == nil {
		return
	}
	for _, arg := range l.List {
		names := make([]string, len(arg.Names))
		for i, name := range arg.Names {
			names[i] = name.Name
		}
		fmt.Fprintf(w, "%s ", strings.Join(names, ", "))
		if err := printer.Fprint(w, fset, arg.Type); err != nil {
			log.Fatal(err)
		}
	}
}

func findDecl(pkg *ast.Package, sym *Sym) *Func {
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			decl, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if matchFunc(sym, decl) {
				return &Func{decl: decl, file: f}
			}
		}
	}
	return nil
}

func matchFunc(sym *Sym, decl *ast.FuncDecl) bool {
	typeName, funcName := sym.Func()
	switch typeName {
	case "":
		return decl.Recv == nil && decl.Name.Name == funcName
	default:
		log.Println(decl.Recv)
		return decl.Recv != nil
	}
}

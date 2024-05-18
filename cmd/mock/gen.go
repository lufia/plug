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
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

type Overlay struct {
	Replace map[string]string
}

type Func struct {
	decl *ast.FuncDecl
	file *ast.File
}

func Generate(sym *Sym) (orig, new string) {
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

	path := fset.File(fn.file.Package).Name()
	file := filepath.Base(path)
	dir := filepath.Join("mock", sym.PkgPath())
	os.MkdirAll(dir, 0755)
	mock := filepath.Join(dir, file)
	w, err := os.Create(mock)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	fn.Replace(w, fset)
	if err := w.Sync(); err != nil {
		log.Fatal(err)
	}
	return path, mock
}

func (fn *Func) Replace(w io.Writer, fset *token.FileSet) {
	name := fn.decl.Name.Name
	fn.decl.Name.Name = "_" + name
	astutil.AddImport(fset, fn.file, "github.com/lufia/mock")
	if err := format.Node(w, fset, fn.file); err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, "func ")
	if fn.decl.Recv != nil {
	}
	fmt.Fprint(w, name)
	fmt.Fprint(w, "(")
	printTypeList(w, fset, fn.decl.Type.Params)
	fmt.Fprint(w, ") (")
	printTypeList(w, fset, fn.decl.Type.Results)
	fmt.Fprintln(w, ") {")
	fmt.Fprintf(w, "\tf := mock.Get(%s)\n", name)
	fmt.Fprintln(w, "\treturn f(d)")
	fmt.Fprintln(w, "}")
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

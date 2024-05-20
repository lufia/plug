package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func ReplaceSyms(f *File, syms []*Sym) (string, error) {
	file := filepath.Base(f.path)
	dir := filepath.Join("mock", f.pkg.path)
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("failed to create %s: %w", dir, err)
	}
	stub := filepath.Join(dir, file)
	w, err := os.Create(stub)
	if err != nil {
		return "", fmt.Errorf("failed to create %s: %w", stub, err)
	}
	defer w.Close()

	if err := rewriteFile(w, f, syms); err != nil {
		return "", fmt.Errorf("failed to rewrite %s: %w", f.path, err)
	}
	if err := w.Sync(); err != nil {
		return "", fmt.Errorf("failed to save a stub: %w", err)
	}
	return stub, nil
}

func rewriteFile(w io.Writer, f *File, syms []*Sym) error {
	astutil.AddImport(f.pkg.fset, f.f, "github.com/lufia/mock")

	var buf bytes.Buffer
	for _, sym := range syms {
		fn := f.pkg.FindFunc(sym)
		if fn == nil {
			log.Fatal("no func")
		}
		fn.Replace(&buf, f.pkg.fset)
	}
	s, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	if err := format.Node(w, f.pkg.fset, f.f); err != nil {
		return err
	}
	fmt.Fprintf(w, "\n%s", s)
	return nil
}

func (fn *Func) Replace(w io.Writer, fset *token.FileSet) {
	name := fn.decl.Name.Name
	fn.decl.Name.Name = "_" + name

	fmt.Fprint(w, "func ")
	if fn.decl.Recv != nil {
		recv := fn.decl.Recv.List[0]
		fmt.Fprint(w, "(%s) ", recvTypeStr(recv.Type))
	}
	fmt.Fprint(w, name)
	if fn.decl.Type.TypeParams != nil {
		fmt.Fprint(w, "[")
		printTypeList(w, fset, fn.decl.Type.TypeParams)
		fmt.Fprint(w, "]")
	}
	fmt.Fprint(w, "(")
	printTypeList(w, fset, fn.decl.Type.Params)
	fmt.Fprint(w, ") (")
	printTypeList(w, fset, fn.decl.Type.Results)
	fmt.Fprintln(w, ") {")
	fmt.Fprintf(w, "\tf := mock.Get(%s, %s)\n", name, fn.decl.Name.Name)
	var args []string
	for _, l := range fn.decl.Type.Params.List {
		names := Map(l.Names, func(i *ast.Ident) string {
			return i.Name
		})
		args = append(args, names...)
	}
	fmt.Fprintf(w, "\treturn f(%s)\n", strings.Join(args, ", "))
	fmt.Fprintln(w, "}")
}

func printTypeList(w io.Writer, fset *token.FileSet, l *ast.FieldList) error {
	if l == nil {
		return nil
	}
	for i, arg := range l.List {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		names := Map(arg.Names, func(i *ast.Ident) string {
			return i.Name
		})
		fmt.Fprintf(w, "%s ", strings.Join(names, ", "))
		if err := printer.Fprint(w, fset, arg.Type); err != nil {
			return err
		}
	}
	return nil
}

func recvTypeStr(expr ast.Expr) string {
	var s string
	for {
		switch p := expr.(type) {
		case *ast.Ident:
			return s + p.Name
		case *ast.UnaryExpr:
			s = "*"
		default:
			panic(p)
		}
	}
}

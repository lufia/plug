package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

type Stub struct { // Plug?
	f   *File
	fns []*Func
}

func Rewrite(stub *Stub) (string, error) {
	filePath := stub.f.path
	name := filepath.Base(filePath)
	dir := filepath.Join("plug", stub.f.pkg.path)
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("failed to create %s: %w", dir, err)
	}
	file := filepath.Join(dir, name)
	w, err := os.Create(file)
	if err != nil {
		return "", fmt.Errorf("failed to create %s: %w", file, err)
	}
	defer w.Close()

	if err := rewriteFile(w, stub); err != nil {
		return "", fmt.Errorf("failed to rewrite %s: %w", filePath, err)
	}
	if err := w.Sync(); err != nil {
		return "", fmt.Errorf("failed to save a stub: %w", err)
	}
	return file, nil
}

func rewriteFile(w io.Writer, stub *Stub) error {
	fset := stub.f.pkg.c.Fset
	astutil.AddImport(fset, stub.f.f, "github.com/lufia/plug")

	var buf bytes.Buffer
	for _, fn := range stub.fns {
		rewriteFunc(&buf, fn)
	}
	s, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("====\n%s\n====\n", s)
	}
	if err := format.Node(w, fset, stub.f.f); err != nil {
		return err
	}
	fmt.Fprintf(w, "\n%s", s)
	return nil
}

func rewriteFunc(w io.Writer, fn *Func) {
	name := fn.fn.Name()
	fn.Rename("_" + name)

	sig := fn.fn.Type().(*types.Signature)
	fmt.Fprint(w, "func ")
	recvName := ""
	if recv := sig.Recv(); recv != nil {
		s := typeStr(recv.Type().Underlying())
		fmt.Fprintf(w, "(%s %s) ", recv.Name(), s)
		recvName = recv.Name() + "."
		// TODO(lufia): sig.RecvTypeParams
	}
	fmt.Fprint(w, name)

	var typeParams []string
	if params := sig.TypeParams(); params != nil {
		fmt.Fprint(w, "[")
		typeParams = printTypeParams(w, params)
		fmt.Fprint(w, "]")
	}

	fmt.Fprint(w, "(")
	paramNames := printVars(w, sig.Params())
	fmt.Fprint(w, ") (")
	resultNames := printVars(w, sig.Results())
	fmt.Fprintln(w, ") {")
	fmt.Fprintln(w, "\tscope := plug.CurrentScope()")
	fmt.Fprintln(w, "\tdefer scope.Delete()")
	if len(typeParams) == 0 {
		fmt.Fprintf(w, "\ts := plug.Func(%q, %s_%s)\n", fn.name, recvName, name)
		fmt.Fprintf(w, "\tf := plug.Get(scope, s, %s_%s, plug.WithParams(map[string]any{\n", recvName, name)
		recordParams(w, sig.Params())
		fmt.Fprintln(w, "\t}))")
	} else {
		s := strings.Join(typeParams, ", ")
		fmt.Fprintf(w, "\ts := plug.Func(%q, %s_%s[%s])\n", fn.name, recvName, name, s)
		fmt.Fprintf(w, "\tf := plug.Get(scope, s, %s_%s[%s], plug.WithParams(map[string]any{\n", recvName, name, s)
		recordParams(w, sig.Params())
		fmt.Fprintln(w, "\t}))")
	}
	if len(resultNames) == 0 {
		fmt.Fprintf(w, "\tf(%s)\n", strings.Join(paramNames, ", "))
	} else {
		fmt.Fprintf(w, "\treturn f(%s)\n", strings.Join(paramNames, ", "))
	}
	fmt.Fprintln(w, "}")
}

func printVars(w io.Writer, vars *types.Tuple) []string {
	if vars == nil {
		return nil
	}
	a := make([]string, vars.Len())
	for i := range vars.Len() {
		v := vars.At(i)
		a[i] = v.Name()
		fmt.Fprintf(w, "%s %s,", v.Name(), typeStr(v.Type()))
	}
	return a
}

func recordParams(w io.Writer, params *types.Tuple) {
	if params == nil {
		return
	}
	for i := range params.Len() {
		v := params.At(i)
		if v.Name() == "_" {
			continue
		}
		fmt.Fprintf(w, "\t\t%[1]q: %[1]s,\n", v.Name())
	}
}

func printTypeParams(w io.Writer, params *types.TypeParamList) []string {
	a := make([]string, params.Len())
	for i := range params.Len() {
		v := params.At(i)
		a[i] = v.Obj().Name()
		fmt.Fprintf(w, "%s %s,", v.Obj().Name(), typeStr(v.Constraint()))
	}
	return a
}

func typeStr(t types.Type) string {
	switch v := t.(type) {
	case *types.Named:
		return types.TypeString(v, types.RelativeTo(v.Obj().Pkg()))
	case *types.Pointer:
		return "*" + typeStr(v.Elem())
	default:
		return t.String()
	}
}

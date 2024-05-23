package main

import (
	"go/ast"
	"go/parser"
	"go/types"
	"slices"

	"golang.org/x/tools/go/loader"
)

type Pkg struct {
	*loader.PackageInfo
	c    *loader.Config
	path string
}

type File struct {
	pkg  *Pkg
	f    *ast.File
	path string
}

type Func struct {
	pkg  *Pkg
	file string
	fn   *types.Func
	name string // pkg/path.Object
}

func LoadPackage(pkgPath string) (*Pkg, error) {
	c := loader.Config{
		ParserMode: parser.ParseComments,
	}
	c.Import(pkgPath)
	p, err := c.Load()
	if err != nil {
		return nil, err
	}
	pkg := p.Package(pkgPath)
	return &Pkg{pkg, &c, pkgPath}, nil
}

func (pkg *Pkg) Lookup(sym *Sym) *Func {
	typeName, funcName := sym.Func()
	var obj types.Object
	switch typeName {
	case "":
		obj = pkg.Pkg.Scope().Lookup(funcName)
	default:
		p := pkg.Pkg.Scope().Lookup(typeName).Type().(*types.Named)
		if p == nil {
			return nil
		}
		for i := range p.NumMethods() {
			m := p.Method(i)
			if m.Name() == funcName {
				obj = m
				break
			}
		}
	}
	fn, ok := obj.(*types.Func)
	if !ok {
		return nil
	}
	f := pkg.c.Fset.File(fn.Pos())
	return &Func{pkg, f.Name(), fn, sym.String()}
}

func (pkg *Pkg) File(name string) *File {
	i := slices.IndexFunc(pkg.Files, func(e *ast.File) bool {
		// e.Name is the name of the package.
		// Thus we should get filename of e from token.FileSet.
		f := pkg.c.Fset.File(e.Pos())
		return f.Name() == name
	})
	if i < 0 {
		return nil
	}
	f := pkg.Files[i]
	return &File{pkg, f, pkg.c.Fset.File(f.Package).Name()}
}

func (fn *Func) File() *File {
	return fn.pkg.File(fn.file)
}

func (fn *Func) Rename(name string) {
	f := fn.File()
	if f == nil {
		panic("unrelated file?")
	}
	for _, d := range f.f.Decls {
		decl, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if f.pkg.Defs[decl.Name] == fn.fn {
			decl.Name.Name = name
			return
		}
	}
}

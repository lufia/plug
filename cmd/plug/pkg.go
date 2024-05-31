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
	c       *loader.Config
	path    string
	version string // If it is empty, maybe it is the stdlib
}

func (pkg *Pkg) PathVersion() string {
	s := pkg.path
	if v := pkg.version; v != "" {
		s += "@" + v
	}
	return s
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

var pkgCache = make(map[string]*Pkg)

func LoadPackage(pkgPath, modVersion string) (*Pkg, error) {
	if pkg, ok := pkgCache[pkgPath]; ok {
		return pkg, nil
	}
	c := loader.Config{
		ParserMode: parser.ParseComments,
	}
	c.Import(pkgPath)
	p, err := c.Load()
	if err != nil {
		return nil, err
	}
	pkg := &Pkg{p.Package(pkgPath), &c, pkgPath, modVersion}
	pkgCache[pkgPath] = pkg
	return pkg, nil
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

var fileCache = make(map[string]*File)

func (pkg *Pkg) File(name string) *File {
	if f, ok := fileCache[name]; ok {
		return f
	}
	i := slices.IndexFunc(pkg.Files, func(e *ast.File) bool {
		// e.Name is the name of the package.
		// Thus we should get filename of e from token.FileSet.
		f := pkg.c.Fset.File(e.Pos())
		return f.Name() == name
	})
	if i < 0 {
		return nil
	}
	fp := pkg.Files[i]
	f := &File{pkg, fp, pkg.c.Fset.File(fp.Package).Name()}
	fileCache[name] = f
	return f
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

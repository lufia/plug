package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
)

type Pkg struct {
	p    *ast.Package
	fset *token.FileSet
	path string
}

type File struct {
	pkg  *Pkg
	f    *ast.File
	path string
}

type Func struct {
	decl *ast.FuncDecl
	f    *File
}

func LoadPackage(pkgPath string) (*Pkg, error) {
	p, err := build.Default.Import(pkgPath, ".", 0)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, p.Dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &Pkg{pkgs[p.Name], fset, pkgPath}, nil
}

func (pkg *Pkg) Files() []*File {
	var a []*File
	for _, f := range pkg.p.Files {
		filePath := pkg.fset.File(f.Package).Name()
		a = append(a, &File{pkg, f, filePath})
	}
	return a
}

func (pkg *Pkg) FindFunc(sym *Sym) *Func {
	for _, f := range pkg.Files() {
		if decl := findFunc(f, sym); decl != nil {
			return &Func{decl, f}
		}
	}
	return nil
}

func findFunc(f *File, sym *Sym) *ast.FuncDecl {
	for _, d := range f.f.Decls {
		decl, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if matchFunc(sym, decl) {
			return decl
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
		return decl.Recv != nil
	}
}

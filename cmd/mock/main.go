package main

import (
	"go/build"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	log.SetFlags(0)

	target, err := build.Default.Import(".", ".", 0)
	if err != nil {
		log.Fatal(err)
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, target.Dir, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	syms := FindMockSetInPkgs(pkgs)
	for _, sym := range syms {
		Generate(sym)
	}
}

func loadPkg(pkgPath string) *build.Package {
	pkg, err := build.Default.Import(pkgPath, ".", 0)
	if err != nil {
		log.Fatal(err)
	}
	return pkg
}

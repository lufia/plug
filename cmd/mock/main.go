package main

import (
	"encoding/json"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
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
	var o Overlay
	o.Replace = make(map[string]string)
	syms := FindMockSetInPkgs(pkgs)
	for _, sym := range syms {
		orig, new := Generate(sym)
		o.Replace[orig] = new
	}
	if err := json.NewEncoder(os.Stdout).Encode(&o); err != nil {
		log.Fatal(err)
	}
}

func loadPkg(pkgPath string) *build.Package {
	pkg, err := build.Default.Import(pkgPath, ".", 0)
	if err != nil {
		log.Fatal(err)
	}
	return pkg
}

package main

import (
	"encoding/json"
	"go/build"
	"log"
	"os"
)

type Overlay struct {
	Replace map[string]string
}

func (o *Overlay) Add(old, new string) {
	if o.Replace == nil {
		o.Replace = make(map[string]string)
	}
	o.Replace[old] = new
}

type Group struct {
	f    *File
	syms []*Sym
}

func (g *Group) Add(sym *Sym) {
	g.syms = append(g.syms, sym)
}

func main() {
	log.SetFlags(0)

	target, err := build.Default.Import(".", ".", 0)
	if err != nil {
		log.Fatal(err)
	}

	syms, err := FindMockSyms(target.Dir)
	if err != nil {
		log.Fatal(err)
	}
	pkgs := GroupSyms(syms)

	var o Overlay
	for _, m := range pkgs {
		for filePath, g := range m {
			s, err := ReplaceSyms(g.f, g.syms)
			if err != nil {
				log.Fatal(err)
			}
			o.Add(filePath, s)
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(&o); err != nil {
		log.Fatal(err)
	}
}

// GroupSyms returns a map of groups indexed pkgPath -> filePath.
func GroupSyms(syms []*Sym) map[string]map[string]*Group {
	pkgs := make(map[string]map[string]*Group)
	for _, sym := range syms {
		pkgPath := sym.PkgPath()
		pkg, err := LoadPackage(pkgPath)
		if err != nil {
			log.Fatal(err)
		}
		fn := pkg.FindFunc(sym)
		if fn == nil {
			log.Fatalf("%s is not exist\n", sym)
		}
		if pkgs[pkgPath] == nil {
			pkgs[pkgPath] = make(map[string]*Group)
		}
		filePath := fn.f.path
		if pkgs[pkgPath][filePath] == nil {
			pkgs[pkgPath][filePath] = &Group{fn.f, nil}
		}
		pkgs[pkgPath][filePath].Add(sym)
	}
	return pkgs
}

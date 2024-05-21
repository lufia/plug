package main

import (
	"encoding/json"
	"flag"
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

var (
	verbose bool
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("mock: ")
	flag.BoolVar(&verbose, "v", false, "enable verbose log")
	flag.Parse()

	target, err := build.Default.Import(".", ".", 0)
	if err != nil {
		log.Fatal(err)
	}

	syms, err := FindMockSyms(target.Dir)
	if err != nil {
		log.Fatal(err)
	}
	stubs := Group(syms)

	var o Overlay
	for filePath, stub := range stubs {
		s, err := Rewrite(stub)
		if err != nil {
			log.Fatal(err)
		}
		o.Add(filePath, s)
	}
	if err := json.NewEncoder(os.Stdout).Encode(&o); err != nil {
		log.Fatal(err)
	}
}

// Group returns a map of Stub indexed by filePath.
func Group(syms []*Sym) map[string]*Stub {
	stubs := make(map[string]*Stub)
	for _, sym := range syms {
		pkgPath := sym.PkgPath()
		pkg, err := LoadPackage(pkgPath)
		if err != nil {
			log.Fatalf("failed to load package %s: %v\n", pkgPath, err)
		}
		fn := pkg.Lookup(sym)
		if fn == nil {
			log.Fatalf("%s is not exist\n", sym)
		}
		f := fn.File()
		stub, ok := stubs[f.path]
		if !ok {
			stub = &Stub{f: f}
			stubs[f.path] = stub
		}
		stub.fns = append(stub.fns, fn)
	}
	return stubs
}

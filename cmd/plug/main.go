package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/mod/modfile"
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
	log.SetPrefix("plug: ")
	flag.BoolVar(&verbose, "v", false, "enable verbose log")
	flag.Parse()

	pkgPath, err := loadPackagePath(".")
	if err != nil {
		log.Fatal(err)
	}
	syms, err := FindPlugSyms(pkgPath)
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

func loadPackagePath(dir string) (string, error) {
	// loader.Import does not handle "." notation that means current package.
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	s := dir
	file := filepath.Join(s, "go.mod")
	for {
		_, err := os.Stat(file)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		up := filepath.Dir(s)
		if up == s {
			return "", fmt.Errorf("go.mod is not exist")
		}
		s = up
		file = filepath.Join(s, "go.mod")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	modPath := modfile.ModulePath(data)
	if modPath == "" {
		return "", fmt.Errorf("%s: invalid go.mod syntax", file)
	}
	slug, err := filepath.Rel(s, dir)
	if err != nil {
		return "", err
	}
	return path.Join(modPath, filepath.ToSlash(slug)), nil
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
			log.Fatalf("%s is not exist or is not exported\n", sym)
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

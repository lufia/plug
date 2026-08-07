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

	pkgPath, modVers, err := loadPackagePath(".")
	if err != nil {
		log.Fatal(err)
	}
	syms, err := FindPlugSyms(pkgPath)
	if err != nil {
		log.Fatal(err)
	}
	stubs := Group(syms, modVers)

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

func loadPackagePath(dir string) (string, map[string]string, error) {
	// loader.Import does not handle "." notation that means current package.
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, err
	}
	s := dir
	file := filepath.Join(s, "go.mod")
	for {
		_, err := os.Stat(file)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", nil, err
		}
		up := filepath.Dir(s)
		if up == s {
			return "", nil, fmt.Errorf("go.mod is not exist")
		}
		s = up
		file = filepath.Join(s, "go.mod")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", nil, err
	}

	f, err := modfile.Parse(file, data, nil)
	if err != nil {
		return "", nil, err
	}
	modPath := f.Module.Mod.Path
	if modPath == "" {
		return "", nil, fmt.Errorf("%s: invalid go.mod syntax", file)
	}
	slug, err := filepath.Rel(s, dir)
	if err != nil {
		return "", nil, err
	}
	pkgPath := path.Join(modPath, filepath.ToSlash(slug))

	modVers := make(map[string]string)
	for _, r := range f.Require {
		modVers[r.Mod.Path] = r.Mod.Version
	}
	return pkgPath, modVers, nil
}

// Group returns a map of Stub indexed by filePath.
func Group(syms []*Sym, modVers map[string]string) map[string]*Stub {
	stubs := make(map[string]*Stub)
	for _, sym := range syms {
		pkgPath := sym.PkgPath()
		pkg, err := LoadPackage(pkgPath, modVers[pkgPath])
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

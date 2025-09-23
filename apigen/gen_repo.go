package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"
)

func GenRepos(tmpl *template.Template, models map[string]ModelDef, dir string) error {
	tmpl = tmpl.Lookup("gen_repo.tmpl")
	for name, md := range models {
		if n, ok := strings.CutPrefix(md.Path, dir); ok {
			mdir := path.Join(dir, n, fmt.Sprintf("%s_repo.go", strings.ToLower(name)))
			f, err := os.Create(mdir)
			if err != nil {
				return err
			}
			defer f.Close()

			md.Extras["Models"] = models
			tmpl.Execute(f, md)
			fmt.Printf("Repo: %s PATH [%s]\n", name, mdir)
		} else {
			log.Fatalf("Repo %s path %s not in dir %s", name, md.Path, dir)
		}
	}
	return nil
}

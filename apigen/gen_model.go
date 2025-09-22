package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"
)

func GenModels(tmpl *template.Template, models map[string]ModelDef, dir string) error {
	tmpl = tmpl.Lookup("gen_model.tmpl")
	for name, md := range models {
		if n, ok := strings.CutPrefix(md.Path, dir); ok {
			mdir := path.Join(dir, n, fmt.Sprintf("%s.go", strings.ToLower(name)))
			f, err := os.Create(mdir)
			if err != nil {
				return err
			}
			defer f.Close()

			err = tmpl.Execute(f, md)
			if err != nil {
				return err
			}
			fmt.Printf("Model: %s PATH [%s]\n", name, mdir)
		} else {
			log.Fatalf("Model %s path %s not in dir %s", name, md.Path, dir)
		}
	}
	return nil
}

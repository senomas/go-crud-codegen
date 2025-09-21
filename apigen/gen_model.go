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

			dfields := make([]map[string]string, 0)
			dmodel := map[string]any{
				"Name": name,
			}
			if n == "" {
				dmodel["Package"] = "model"
			} else {
				dmodel["Package"] = path.Base(n)
			}

			for _, fd := range md.Fields {
				df := map[string]string{
					"Model": name,
					"Name":  fd.ID,
					"Field": fd.Field,
					"Type":  fd.GoType,
				}
				dfields = append(dfields, df)
			}
			dmodel["Fields"] = dfields

			err = tmpl.Execute(f, dmodel)
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

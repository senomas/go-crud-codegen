package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"
)

func GenModels(models map[string]ModelDef, dir string) error {
	tmpl, err := template.New("model").Parse(templ_model)
	if err != nil {
		return err
	}
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
					"BT":    "`",
					"Name":  fd.ID,
					"Field": fd.Field,
					"Type":  fd.GoType,
				}
				dfields = append(dfields, df)
			}
			dmodel["Fields"] = dfields

			tmpl.Execute(f, dmodel)
			fmt.Printf("Model: %s PATH [%s]\n", name, mdir)
		} else {
			log.Fatalf("Model %s path %s not in dir %s", name, md.Path, dir)
		}
	}
	return nil
}

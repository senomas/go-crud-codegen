package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

func main() {
	models := make(map[string]ModelDef)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dir = path.Join(dir, "../api/model")
	fmt.Printf("path %s\n", dir)
	err = LoadModels(models, dir)
	if err != nil {
		panic(err)
	}
	tmpl := Templates()

	tmpls := map[string]*template.Template{
		"migration|.sql": tmpl.Lookup("gen_sql.tmpl"),
		".go":            tmpl.Lookup("gen_model.tmpl"),
		"_repo.go":       tmpl.Lookup("gen_repo.tmpl"),
	}
	for name, md := range models {
		if md.Extras == nil {
			md.Extras = make(map[string]any)
		}
		md.Extras["model"] = func(id string) (*ModelDef, error) {
			if m, ok := models[id]; ok {
				return &m, nil
			}
			return nil, fmt.Errorf("model %s not found", id)
		}
		if n, ok := strings.CutPrefix(md.Path, dir); ok {
			for k, t := range tmpls {
				var fn string
				kk := strings.SplitN(k, "|", 2)
				if len(kk) == 1 {
					fn = path.Join(dir, n, fmt.Sprintf("%s%s", strings.ToLower(name), k))
				} else {
					if kk[0] == "migration" {
						if mig, ok := md.Extras["mutation"].(string); ok && mig != "" {
							fn = path.Join(dir, n, "..", kk[0], fmt.Sprintf("%s-%s%s", mig, strings.ToLower(name), kk[1]))
						} else {
							log.Fatalf("md.Extras['mutation'] is not string or empty: %T", md.Extras["mutation"])
						}
					} else {
						fn = path.Join(dir, n, "..", kk[0], fmt.Sprintf("%s%s", strings.ToLower(name), kk[1]))
					}
				}
				fmt.Printf("Generating %s ...\n", fn)
				f, err := os.Create(fn)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()

				err = t.Execute(f, md)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

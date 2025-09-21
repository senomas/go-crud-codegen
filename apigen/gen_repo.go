package main

import (
	"fmt"
	"html/template"
	"log"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
)

func GenRepos(models map[string]ModelDef, dir string) error {
	tmpl, err := template.New("repo").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(templ_repo)
	if err != nil {
		return err
	}

	for name, md := range models {
		if n, ok := strings.CutPrefix(md.Path, dir); ok {
			mdir := path.Join(dir, n, fmt.Sprintf("%s_repo.go", strings.ToLower(name)))
			f, err := os.Create(mdir)
			if err != nil {
				return err
			}
			defer f.Close()

			dmodel := map[string]any{
				"BT":    "`",
				"Name":  name,
				"Table": md.Table,
			}
			if n == "" {
				dmodel["Package"] = "model"
			} else {
				dmodel["Package"] = path.Base(n)
			}

			fields := make([]map[string]any, 0)
			pkeys := make([]map[string]any, 0)
			updatables := make([]map[string]any, 0)
			for _, fd := range md.Fields {
				pkey := slices.Contains(md.PKey, fd.ID)
				df := map[string]any{
					"BT":     "`",
					"Model":  name,
					"Name":   fd.ID,
					"Field":  fd.Field,
					"Type":   fd.Type,
					"GoType": fd.GoType, "PKey": pkey,
					"Filtertype": fd.Type == "text" || fd.Type == "autoincrement",
				}
				switch fd.Type {
				case "text":
					df["Filter"] = "text"
				case "autoincrement", "int", "float", "numeric":
					df["Filter"] = "numeric"
				}
				fields = append(fields, df)
				if pkey {
					pkeys = append(pkeys, maps.Clone(df))
				} else if fd.Type != "password" && fd.Type != "salt" {
					updatables = append(updatables, maps.Clone(df))
				}
			}
			fields[len(fields)-1]["Last"] = true
			pkeys[len(pkeys)-1]["Last"] = true
			updatables[len(updatables)-1]["Last"] = true
			dmodel["Fields"] = fields
			dmodel["PKeys"] = pkeys
			dmodel["Updatables"] = updatables

			tmpl.Execute(f, dmodel)
			fmt.Printf("Repo: %s PATH [%s]\n", name, mdir)
		} else {
			log.Fatalf("Repo %s path %s not in dir %s", name, md.Path, dir)
		}
	}
	return nil
}

package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"slices"
)

func Templates() *template.Template {
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"dict": func(kv ...any) (map[string]any, error) {
			if len(kv)%2 != 0 {
				return nil, fmt.Errorf("dict requires even args")
			}
			m := map[string]any{}
			for i := 0; i < len(kv); i += 2 {
				k, ok := kv[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[k] = kv[i+1]
			}
			return m, nil
		},
		"snakeCase": toSnakeCase,
		"inSlices": func(v any, k string) bool {
			if s, ok := v.([]string); ok {
				return slices.Contains(s, k)
			} else if s, ok := v.([]any); ok {
				for _, sv := range s {
					if sv == k {
						return true
					}
				}
				return false
			}
			fmt.Printf("inSlices: v is not []string: %T\n", v)
			return false
		},
		"goType": func(fo FieldDef) string {
			vt := fo.Type
			switch fo.Type {
			case "autoincrement":
				vt = "int64"
			case "text":
				vt = "string"
			case "password", "salt", "secret":
				vt = "string"
			case "many-to-one":
				vt = "*" + fo.Extras["ref"].(string)
			}
			if fo.Null {
				vt = "*" + vt
			}
			return vt
		},
		"goSqlNullType": func(fo FieldDef) string {
			vt := fo.Type
			switch fo.Type {
			case "autoincrement":
				vt = "sql.NullInt64"
			case "text":
				vt = "sql.NullString"
			case "password", "salt", "secret":
				vt = "sql.NullString"
			case "many-to-one":
				vt = "*" + fo.Extras["ref"].(string)
			}
			if fo.Null {
				vt = "*" + vt
			}
			return vt
		},
		"isPk": func(field FieldDef, model ModelDef) bool {
			return slices.Contains(model.PKeys, field.ID)
		},
		"isUpdatable": func(field FieldDef, model ModelDef) bool {
			return !slices.Contains(model.PKeys, field.ID) && field.Type != "password" && field.Type != "salt"
		},
	}).ParseGlob("*.tmpl")
	if err != nil {
		panic(err)
	}
	return tmpl
}

func main() {
	models := make(map[string]ModelDef)
	sp, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	sp = path.Join(sp, "../api/model")
	fmt.Printf("path %s\n", sp)
	err = LoadModels(models, sp)
	if err != nil {
		panic(err)
	}
	tmpl := Templates()

	err = GenModels(tmpl, models, sp)
	if err != nil {
		panic(err)
	}
	err = GenRepos(tmpl, models, sp)
	if err != nil {
		panic(err)
	}
}

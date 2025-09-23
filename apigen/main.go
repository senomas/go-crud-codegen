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
		"model": func(m ModelDef, obj any) ModelDef {
			if md, ok := obj.(ModelDef); ok {
				return md
			} else if fd, ok := obj.(FieldDef); ok {
				if ref, ok := fd.Extras["ref"].(string); ok {
					if models, ok := m.Extras["Models"].(map[string]ModelDef); ok {
						if md, ok := models[ref]; ok {
							return md
						}
					} else {
						log.Fatalf("m.Extras['Models'] is not map[string]ModelDef: %T", m.Extras["Models"])
					}
				}
			} else if ref, ok := obj.(string); ok {
				if models, ok := m.Extras["Models"].(map[string]ModelDef); ok {
					if md, ok := models[ref]; ok {
						return md
					}
				} else {
					log.Fatalf("m.Extras['Models'] is not map[string]ModelDef: %T", m.Extras["Models"])
				}
			}
			log.Fatalf("model %+v not found in models", obj)
			return ModelDef{}
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
			if fo.Null {
				switch fo.Type {
				case "autoincrement":
					vt = "sql.NullInt64"
				case "version":
					vt = "sql.NullInt64"
				case "text":
					vt = "sql.NullString"
				case "password", "salt", "secret":
					vt = "sql.NullString"
				case "timestamp":
					vt = "time.Time"
				case "many-to-one":
					vt = "*" + fo.Extras["ref"].(string)
				}
			} else {
				switch fo.Type {
				case "autoincrement":
					vt = "int64"
				case "version":
					vt = "int64"
				case "text":
					vt = "string"
				case "password", "salt", "secret":
					vt = "string"
				case "timestamp":
					vt = "time.Time"
				case "many-to-one":
					vt = "*" + fo.Extras["ref"].(string)
				}
			}
			return vt
		},
		"goSqlNullType": func(fo FieldDef) string {
			vt := fo.Type
			switch fo.Type {
			case "autoincrement":
				vt = "sql.NullInt64"
			case "ver":
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
		"goSqlNullValue": func(fo FieldDef) string {
			vt := fo.Type
			switch fo.Type {
			case "autoincrement":
				vt = ".Int64"
			case "version":
				vt = ".Int64"
			case "text":
				vt = ".String"
			case "password", "salt", "secret":
				vt = ".String"
			case "many-to-one":
				vt = ""
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
			return !slices.Contains(model.PKeys, field.ID) && field.Type != "password" && field.Type != "salt" && field.Type != "version"
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

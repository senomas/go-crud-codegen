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
		"inSlices": func(pkeys []string, id string) bool {
			return slices.Contains(pkeys, id)
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

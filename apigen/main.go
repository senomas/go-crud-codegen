package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
)

func Templates() *template.Template {
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
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

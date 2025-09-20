package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

func main() {
	models := make(map[string]ModelDef)
	sp, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	sp = path.Join(sp, "../api")
	fmt.Printf("path %s\n", sp)
	err = LoadModels(models, sp)
	if err != nil {
		panic(err)
	}
	err = GenModels(models, sp)
	if err != nil {
		panic(err)
	}
	err = GenRepos(models, sp)
	if err != nil {
		panic(err)
	}
}

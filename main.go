package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"
)

func main() {
	models := make(map[string]ModelDef)
	dialect := os.Args[1]
	module := os.Args[2]
	err := LoadModels(models, "/work/app/model", module)
	if err != nil {
		panic(err)
	}
	tmpl := Templates("/work/codegen", dialect)

	tmpls := []*template.Template{
		tmpl.Lookup("gen_sql.tmpl"),
		tmpl.Lookup("gen_model.tmpl"),
		tmpl.Lookup("gen_store.tmpl"),
		tmpl.Lookup("gen_handler.tmpl"),
	}
	files := []string{}
	{
		t := tmpl.Lookup("gen_model_root.tmpl")
		var bb bytes.Buffer
		err = t.Execute(&bb, models)
		if err != nil {
			log.Fatalf("Executing template for root %v", err)
		}

		files = writeOut(files, &bb)
	}
	for name, md := range models {
		for _, t := range tmpls {
			var bb bytes.Buffer
			err = t.Execute(&bb, &md)
			if err != nil {
				log.Fatalf("Executing template for model %s: %v", name, err)
			}

			files = writeOut(files, &bb)
		}
	}
	eargs := []string{"-w"}
	for _, f := range files {
		if strings.HasSuffix(f, ".go") {
			eargs = append(eargs, f)
		}
	}
	out, err := exec.Command("goimports", eargs...).CombinedOutput()
	if err != nil {
		log.Fatalf("goimports error: %v\n%s", err, out)
	}
	out, err = exec.Command("gofmt", eargs...).CombinedOutput()
	if err != nil {
		log.Fatalf("gofmt error: %v\n%s", err, out)
	}
	for _, fn := range files {
		diff(fn)
	}
}

var reFile = regexp.MustCompile(`^(?://|--)\s+FILE(?:-([0-9a-fA-F]+))?:\s*(.+)$`)

func writeOut(files []string, bb *bytes.Buffer) []string {
	var err error
	scanner := bufio.NewScanner(bb)
	if scanner.Scan() {
		ln := scanner.Text()
		m := reFile.FindStringSubmatch(ln)
		if m != nil {
			if m[1] == "" {
				fmt.Printf("Generating file %s\n", m[2])
				files = append(files, m[2])
				f, err := os.Create(path.Join("/work/app", m[2]))
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()

				for scanner.Scan() {
					f.WriteString(scanner.Text() + "\n")
				}
			} else {
				files = append(files, m[2])
				fmt.Printf("Generating file %s\n", m[2])
				f := []*os.File{nil}
				f[0], err = os.Create(path.Join("/work/app", m[2]))
				if err != nil {
					log.Fatal(err)
				}
				defer f[0].Close()

				for scanner.Scan() {
					ln = scanner.Text()
					ms := reFile.FindStringSubmatch(ln)
					if ms != nil && ms[1] == m[1] {
						f[0].Close()

						files = append(files, ms[2])
						fmt.Printf("Generating file %s\n", ms[2])
						f[0], err = os.Create(path.Join("/work/app", ms[2]))
						if err != nil {
							log.Fatal(err)
						}
					} else {
						f[0].WriteString(scanner.Text() + "\n")
					}
				}
			}
		} else {
			log.Fatalf("First line of generated content does not specify file name")
		}
	}
	return files
}

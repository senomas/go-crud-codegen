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
	args := os.Args
	fmt.Printf("crudgen %+v\n", args)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stat(path.Join(home, "go/bin/goimports")); os.IsNotExist(err) {
		out, err := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest").CombinedOutput()
		if err != nil {
			log.Fatalf("go install goimports error: %v\n%s", err, out)
		}
	}
	if _, err = os.Stat(path.Join(home, "go/bin/goimports")); os.IsNotExist(err) {
		log.Fatalf("goimports not found at %s", path.Join(home, "go/bin/goimports"))
	}
	fmt.Printf("Working directory: %s\n", dir)
	err = os.Chdir(args[1])
	if err != nil {
		log.Fatal(err)
	}
	gitStatus()
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}
	models := make(map[string]ModelDef)
	dir = path.Join(dir, "./model")
	err = LoadModels(models, dir, args[3])
	if err != nil {
		panic(err)
	}
	dialect := args[2]
	tmpl := Templates(dialect)

	tmpls := []*template.Template{
		tmpl.Lookup("gen_sql.tmpl"),
		tmpl.Lookup("gen_model.tmpl"),
		tmpl.Lookup("gen_store.tmpl"),
		tmpl.Lookup("gen_handler.tmpl"),
	}
	files := []string{}
	for name, md := range models {
		md.model = func(id string) (*ModelDef, error) {
			if m, ok := models[id]; ok {
				return &m, nil
			}
			return nil, fmt.Errorf("model %s not found", id)
		}
		re := regexp.MustCompile(`^(?://|--)\s+FILE(?:-([0-9a-fA-F]+))?:\s*(.+)$`)
		for _, t := range tmpls {
			var bb bytes.Buffer
			err = t.Execute(&bb, &md)
			if err != nil {
				log.Fatalf("Executing template for model %s: %v", name, err)
			}

			scanner := bufio.NewScanner(&bb)
			if scanner.Scan() {
				ln := scanner.Text()
				m := re.FindStringSubmatch(ln)
				if m != nil {
					if m[1] == "" {
						fmt.Printf("Generating file %s\n", m[2])
						files = append(files, m[2])
						f, err := os.Create(path.Join(args[1], m[2]))
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
						f[0], err = os.Create(path.Join(args[1], m[2]))
						if err != nil {
							log.Fatal(err)
						}
						defer f[0].Close()

						for scanner.Scan() {
							ln = scanner.Text()
							ms := re.FindStringSubmatch(ln)
							if ms != nil && ms[1] == m[1] {
								f[0].Close()

								files = append(files, ms[2])
								fmt.Printf("Generating file %s\n", ms[2])
								f[0], err = os.Create(path.Join(args[1], ms[2]))
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
		}
	}
	err = os.Chdir(args[1])
	if err != nil {
		log.Fatal(err)
	}
	eargs := []string{"-w"}
	for _, f := range files {
		if strings.HasSuffix(f, ".go") {
			eargs = append(eargs, f)
		}
	}
	if _, err = os.Stat(path.Join(home, "go/bin/goimports")); os.IsNotExist(err) {
		out, err := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest").CombinedOutput()
		if err != nil {
			log.Fatalf("go install goimports error: %v\n%s", err, out)
		}
	}
	out, err := exec.Command(path.Join(home, "go/bin/goimports"), eargs...).CombinedOutput()
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

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

type FieldDef struct {
	ID     string         `yaml:"id"`
	Field  string         `yaml:"field,omitempty"`
	Type   string         `yaml:"type"`
	Null   bool           `yaml:"nullable,omitempty"`
	Extras map[string]any `yaml:",inline"`
}

type UniqueDef struct {
	ID     string         `yaml:"id"`
	Fields []string       `yaml:"fields,omitempty"`
	Extras map[string]any `yaml:",inline"`
}

type ModelDef struct {
	Table   string         `yaml:"table"`
	Extras  map[string]any `yaml:",inline"`
	Fields  []FieldDef     `yaml:"fields"`
	PKeys   []string       `yaml:"pkeys"`
	Uniques []UniqueDef    `yaml:"uniques"`
	ID      string
	Path    string
	Package string
}

func LoadModels(models map[string]ModelDef, dir string) error {
	fmt.Printf("LoadModels %s\n", dir)
	err := filepath.WalkDir(dir, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name != dir {
				return LoadModels(models, name)
			}
			return nil
		}
		if filepath.Ext(name) == ".yml" || filepath.Ext(name) == ".yaml" {
			fmt.Printf("Load file %s\n", name)
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}

			m := make(map[string]ModelDef)
			if err := yaml.Unmarshal(data, &m); err != nil {
				return fmt.Errorf("unmarshal %s: %w", name, err)
			}
			for n, mo := range m {
				mo.Path = path.Dir(name)
				mo.ID = n
				if n == "" {
					mo.Package = "model"
				} else {
					mo.Package = path.Base(dir)
				}
				for i := range mo.Fields {
					if mo.Fields[i].Field == "" {
						mo.Fields[i].Field = toSnakeCase(mo.Fields[i].ID)
					}
				}
				models[n] = mo
			}
		}
		return nil
	})
	return err
}

func toSnakeCase(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4) // small buffer growth for underscores

	for i, r := range s {
		if unicode.IsUpper(r) {
			// add underscore before uppercase if not first char
			// and previous char isn't underscore
			if i > 0 && s[i-1] != '_' {
				// also avoid double underscore if next is uppercase too
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

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
	Filter *bool          `yaml:"filter,omitempty"`
	Sorted *bool          `yaml:"sorted,omitempty"`
	Extras map[string]any `yaml:",inline"`
	GoType string
}

type ModelDef struct {
	Path    string
	Table   string         `yaml:"table"`
	Audit   bool           `yaml:"audit"`
	Version bool           `yaml:"version"`
	Extras  map[string]any `yaml:",inline"`
	Fields  []FieldDef     `yaml:"fields"`
	PKey    []string       `yaml:"pkey"`
	Uniques []string       `yaml:"uniques"`
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
				for i := range mo.Fields {
					if mo.Fields[i].Field == "" {
						mo.Fields[i].Field = toSnakeCase(mo.Fields[i].ID)
					}
					switch mo.Fields[i].Type {
					case "autoincrement":
						mo.Fields[i].GoType = "int64"
					case "text":
						mo.Fields[i].GoType = "string"
					case "password", "salt":
						mo.Fields[i].GoType = "string"
						v := false
						mo.Fields[i].Sorted = &v
						mo.Fields[i].Filter = &v
					case "many-to-one":
						mo.Fields[i].GoType = "*" + mo.Fields[i].Extras["ref"].(string)
					default:
						mo.Fields[i].GoType = "any"
					}
					if mo.Fields[i].Null {
						mo.Fields[i].GoType = "*" + mo.Fields[i].GoType
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

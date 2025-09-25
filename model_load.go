package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
					mo.Fields[i].model = func() *ModelDef {
						return &mo
					}
				}
				for i := range mo.Uniques {
					mo.Uniques[i].Model = func() *ModelDef {
						return &mo
					}
				}
				mo.model = func(id string) (*ModelDef, error) {
					for _, m := range models {
						if m.ID == id {
							return &m, nil
						}
					}
					return nil, fmt.Errorf("model %s not found", id)
				}
				models[n] = mo
			}
		}
		return nil
	})
	return err
}

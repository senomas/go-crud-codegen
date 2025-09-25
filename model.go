package main

import (
	"fmt"
	"slices"
)

type ModelDef struct {
	Table   string      `yaml:"table"`
	Fields  []FieldDef  `yaml:"fields"`
	CPKeys  []string    `yaml:"pkeys"`
	Uniques []UniqueDef `yaml:"uniques"`
	Seq     int         `yaml:"seq,omitempty"`
	ID      string
	Path    string
	Package string
	model   func(id string) (*ModelDef, error)
}

func (m *ModelDef) PKeys() []*FieldDef {
	res := []*FieldDef{}
	for _, f := range m.Fields {
		if slices.Contains(m.CPKeys, f.ID) {
			res = append(res, &f)
		}
	}
	return res
}

func (m *ModelDef) Field(id string) (*FieldDef, error) {
	for _, f := range m.Fields {
		if f.ID == id {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("field %s not found", id)
}

type FK struct {
	ID    string `yaml:"id"`
	Field string `yaml:"field"`
}

type FieldDef struct {
	ID         string   `yaml:"id"`
	Field      string   `yaml:"field,omitempty"`
	Type       string   `yaml:"type"`
	Null       bool     `yaml:"nullable,omitempty"`
	Ref        string   `yaml:"ref"`
	CRefKeys   []FK     `yaml:"refKeys"`
	CRefFields []string `yaml:"refFields"`
	Seq        int      `yaml:"seq,omitempty"`
	MapTable   string   `yaml:"mapTable,omitempty"`
	CMapKeys   []FK     `yaml:"mapKeys"`
	model      func() *ModelDef
}

func (f *FieldDef) Model() *ModelDef {
	return f.model()
}

func (f *FieldDef) RefModel() (*ModelDef, error) {
	return f.model().model(f.Ref)
}

type FieldRef struct {
	ID    string
	Field string
	Ref   *FieldDef
}

type UniqueDef struct {
	ID      string   `yaml:"id"`
	CFields []string `yaml:"fields,omitempty"`
	Model   func() *ModelDef
}

func (u *UniqueDef) Fields() []*FieldDef {
	res := []*FieldDef{}
	for _, fid := range u.CFields {
		m := u.Model()
		for _, f := range m.Fields {
			if f.ID == fid {
				res = append(res, &f)
			}
		}
	}
	return res
}

package main

import (
	"fmt"
	"slices"
	"text/template"
	"time"
)

func Templates() *template.Template {
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"now": func() string {
			return time.Now().In(time.Local).Format(time.RFC3339)
		},
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
		"model": func(m ModelDef, obj any) (*ModelDef, error) {
			if fd, ok := obj.(FieldDef); ok {
				if ref, ok := fd.Extras["ref"].(string); ok {
					if fmod, ok := m.Extras["model"].(func(string) (*ModelDef, error)); ok {
						return fmod(ref)
					}
					return nil, fmt.Errorf("m.Extras['model'] is not func(string) (*ModelDef, error): %T", m.Extras["model"])
				}
			} else {
				return nil, fmt.Errorf("obj is not FieldDef: %T", obj)
			}
			return nil, fmt.Errorf("model %+v not found in models", obj)
		},
		"refKeys":   refKeys,
		"refFields": refFields,
		"mapSelf":   mapSelf,
		"mapRef":    mapRef,
		"mapFields": mapFields,
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
				case "many-to-many":
					vt = "[]" + fo.Extras["ref"].(string)
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
				case "many-to-many":
					vt = "[]" + fo.Extras["ref"].(string)
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
				vt = "Int64"
			case "version":
				vt = "Int64"
			case "text":
				vt = "String"
			case "password", "salt", "secret":
				vt = "String"
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

package main

import (
	"fmt"
	"maps"
	"path"
	"strings"
	"text/template"
	"time"
	"unicode"
)

func Templates(dir, dialect string) *template.Template {
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
		"modelArgs": func(m ModelDef, key string, value any) ModelDef {
			mf := m // copy
			if m.Args == nil {
				mf.Args = map[string]any{}
			} else {
				mf.Args = maps.Clone(m.Args)
			}
			mf.Args[key] = value
			return mf
		},
		"fieldArgs": func(field FieldDef, key string, value any) FieldDef {
			nf := field // copy
			if field.Args == nil {
				nf.Args = map[string]any{}
			} else {
				nf.Args = maps.Clone(field.Args)
			}
			nf.Args[key] = value
			return nf
		},
		"snakeCase": toSnakeCase,
	}).ParseGlob(path.Join(dir, "base", "*.tmpl"))
	if err != nil {
		panic(err)
	}
	tmpl, err = tmpl.ParseGlob(path.Join(dir, dialect, "*.tmpl"))
	if err != nil {
		panic(err)
	}
	return tmpl
}

func toSnakeCase(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && s[i-1] != '_' {
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

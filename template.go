package main

import (
	"fmt"
	"path"
	"strings"
	"text/template"
	"time"
	"unicode"
)

func Templates(dialect string) *template.Template {
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
		"snakeCase": toSnakeCase,
	}).ParseGlob(path.Join("base", "*.tmpl"))
	if err != nil {
		panic(err)
	}
	tmpl, err = tmpl.ParseGlob(path.Join(dialect, "*.tmpl"))
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

package main

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func xxTestRepo(t *testing.T) {
	t.Run("QueryFindCount", func(t *testing.T) {
		tmpl, err := template.New("repo").Funcs(template.FuncMap{
			"add": func(a, b int) int { return a + b },
		}).Parse(templ_repo)
		assert.NoError(t, err)

		var buf bytes.Buffer
		tmpl.ExecuteTemplate(&buf, "QueryFindCount", map[string]any{
			"BT":    "`",
			"Name":  "User",
			"Table": "app_usr",
			"PKeys": []map[string]any{
				{
					"Field": "id",
					"Name":  "ID",
				},
			},
		})
		assert.Equal(t, "SELECT COUNT(id) FROM app_usr", buf.String())
	})
}

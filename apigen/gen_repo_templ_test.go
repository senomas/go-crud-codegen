package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepo(t *testing.T) {
	tmpl := Templates()
	t.Run("QueryFindCount", func(t *testing.T) {
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

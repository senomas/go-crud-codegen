package model_test

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"hanoman.co.id/mwui/api/model"
)

func TestCrud(t *testing.T) {
	ctx := context.Background()
	t.Run("Create user Admin", func(t *testing.T) {
		repos, ctx, err := model.GetRepos("sqlite", ctx)
		assert.NoError(t, err)
		assert.NotNil(t, repos)
		assert.NotNil(t, ctx)
		user := model.User{
			Email: "admin@example.com",
			Name:  "Admin",
		}

		id, err := repos.User().Create(user)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), id)
	})

	t.Run("Get Admin", func(t *testing.T) {
		repos, ctx, err := model.GetRepos("sqlite", ctx)
		assert.NoError(t, err)
		assert.NotNil(t, repos)
		assert.NotNil(t, ctx)
		user, err := repos.User().Get(1)
		assert.NoError(t, err)

		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.Equal(t, "", user.Salt)
		assert.Equal(t, "", user.Password)
		assert.Equal(t, "", user.CreatedAt)
		log.Print(user)
	})
}

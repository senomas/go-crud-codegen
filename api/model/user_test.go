package model_test

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"hanoman.co.id/mwui/api/model"

	_ "github.com/mattn/go-sqlite3"
)

func TestUserCrud(t *testing.T) {
	dsn := "file:../app.db?_busy_timeout=5000&cache=shared&mode=rwc&_foreign_keys=on"
	db, err := sql.Open("sqlite3", dsn)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	userRepo := model.NewUserRepository(context.Background(), db)

	t.Run("Create user Admin", func(t *testing.T) {
		user := model.User{
			Email: "admin@example.com",
			Name:  "Admin",
		}

		res, err := userRepo.Create(user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.ID)
	})

	t.Run("Get Admin", func(t *testing.T) {
		user, err := userRepo.Get(1)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.Equal(t, "", user.Salt)
		assert.Equal(t, "", user.Password)
		log.Print(user)
	})
}

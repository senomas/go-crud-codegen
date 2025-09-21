package model_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"hanoman.co.id/mwui/api/model"
)

func TestUserCrud(t *testing.T) {
	repos := model.GetRepos()
	ctx := context.Background()

	t.Run("Create user Admin", func(t *testing.T) {
		user := model.User{
			Email: "admin@example.com",
			Name:  "Admin",
		}

		res, err := repos.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.ID)
	})

	t.Run("Get user Admin", func(t *testing.T) {
		user, err := repos.User().Get(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.Equal(t, "", user.Salt)
		assert.Equal(t, "", user.Password)
	})

	t.Run("Update user Admin", func(t *testing.T) {
		user := model.User{
			ID:    1,
			Email: "admin@demo.com",
			Name:  "Admin",
		}

		err := repos.User().Update(ctx, user)
		assert.NoError(t, err)
	})

	t.Run("Get user Admin after update", func(t *testing.T) {
		user, err := repos.User().Get(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@demo.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.Equal(t, "", user.Salt)
		assert.Equal(t, "", user.Password)
	})

	t.Run("Find all user", func(t *testing.T) {
		users, total, err := repos.User().Find(ctx, nil, nil, 10, 0)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), total, "total must 1")
		assert.Equal(t, 1, len(users), "len(users) must 1")
		assert.Equal(t, "admin@demo.com", users[0].Email)
		assert.Equal(t, "Admin", users[0].Name)
		assert.Equal(t, "", users[0].Salt)
		assert.Equal(t, "", users[0].Password)
	})

	t.Run("Create user Staff", func(t *testing.T) {
		user := model.User{
			Email: "staff@demo.com",
			Name:  "Staff",
		}

		res, err := repos.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(2), res.ID)
	})

	t.Run("Get user Staff", func(t *testing.T) {
		user, err := repos.User().Get(ctx, 2)
		assert.NoError(t, err)

		assert.Equal(t, int64(2), user.ID)
		assert.Equal(t, "staff@demo.com", user.Email)
		assert.Equal(t, "Staff", user.Name)
		assert.Equal(t, "", user.Salt)
		assert.Equal(t, "", user.Password)
	})

	t.Run("Create user Operator 1", func(t *testing.T) {
		user := model.User{
			Email: "opr1@demo.com",
			Name:  "Operator 1",
		}

		res, err := repos.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(3), res.ID)
	})

	t.Run("Create user Operator 2", func(t *testing.T) {
		user := model.User{
			Email: "opr2@demo.com",
			Name:  "Operator 2",
		}

		res, err := repos.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(4), res.ID)
	})

	t.Run("Create 23 dummy user", func(t *testing.T) {
		for i := 1; i <= 23; i++ {
			user := model.User{
				Email: fmt.Sprintf("dummy%d@demo.com", i),
				Name:  fmt.Sprintf("Dummy %d", i),
			}

			res, err := repos.User().Create(ctx, user)
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, int64(4+i), res.ID)
		}
	})

	t.Run("Find all user", func(t *testing.T) {
		users, total, err := repos.User().Find(ctx, nil, nil, 10, 0)
		assert.NoError(t, err)

		assert.Equal(t, int64(27), total, "total must match")
		assert.Equal(t, 10, len(users), "len(users) must match")
		for i, u := range []string{"admin@demo.com", "staff@demo.com", "opr1@demo.com", "opr2@demo.com", "dummy1@demo.com"} {
			assert.Equal(t, u, users[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("Find all user limit 5 offset 5", func(t *testing.T) {
		users, total, err := repos.User().Find(ctx, nil, nil, 5, 5)
		assert.NoError(t, err)

		assert.Equal(t, int64(27), total, "total must match")
		assert.Equal(t, 5, len(users), "len(users) must match")
		for i, u := range []string{"dummy2@demo.com", "dummy3@demo.com", "dummy4@demo.com"} {
			assert.Equal(t, u, users[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("Find users like dummy% limit 5", func(t *testing.T) {
		users, total, err := repos.User().Find(ctx, []model.UserFilter{{
			Field: model.UserField_Name,
			Op:    model.FilterOp_Like,
			Value: "dummy%",
		}}, nil, 5, 0)
		assert.NoError(t, err)

		assert.Equal(t, int64(5), total, "total must match")
		assert.Equal(t, 5, len(users), "len(users) must match")
	})
}

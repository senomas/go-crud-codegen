package model_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hanoman.co.id/crudgen/model"
)

func TestUserCrud(t *testing.T) {
	store := model.GetStore()

	ctx := context.Background()
	createTime := time.Now().Add(-time.Hour * 24)
	createTime1 := time.Now().Add(-time.Hour * 24).Add(1 * time.Second)
	updateTime := time.Now().Add(-time.Hour * 24).Add(5 * time.Minute)

	t.Run("Create role Admin", func(t *testing.T) {
		role := model.Role{
			Name: "Admin",
			Privileges: `{
				"role":{"read":true,"create":true,"delete":true,"update":true},
				"user":{"read":true,"create":true,"delete":true,"update":true},
				"audit":{"read":true},
				"param":{"read":true,"create":true,"delete":true,"update":true},
				"identity":{"read":true,"create":true,"delete":true,"update":true},
				"tcpgwAccess":{"read":true,"create":true,"delete":true,"update":true},
				"tcpgwAccessIP":{"read":true,"create":true,"delete":true,"update":true},
				"kaltimtara":{"read":true,"create":false,"delete":false,"update":true}
			}`,
		}

		res, err := store.Role().Create(ctx, role)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.ID)
	})

	t.Run("Create role Opr", func(t *testing.T) {
		role := model.Role{
			Name: "Opr",
		}

		res, err := store.Role().Create(ctx, role)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(2), res.ID)
	})

	t.Run("Create role Staff", func(t *testing.T) {
		role := model.Role{
			Name: "Staf",
		}

		res, err := store.Role().Create(ctx, role)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(3), res.ID)
	})

	t.Run("Create user Admin", func(t *testing.T) {
		user := model.User{
			Version:   1,
			Email:     "admin@example.com",
			Name:      "Admin",
			Password:  "admin123",
			Roles:     []model.Role{{ID: 1}, {ID: 2}},
			CreatedAt: createTime,
		}

		res, err := store.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.ID)
	})

	var hpass string
	var version int64
	t.Run("Get user Admin", func(t *testing.T) {
		user, err := store.User().Get(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.NotEqual(t, "", user.Password)
		assert.NotEqual(t, "admin123", user.Password)
		fmt.Printf("hashed password %s\n", user.Password)

		assert.Equal(t, 2, len(user.Roles))
		assert.Equal(t, "Admin", user.Roles[0].Name)
		assert.Equal(t, int64(2), user.Roles[1].ID)
		assert.Equal(t, "Opr", user.Roles[1].Name)

		check, err := user.VerifyPassword("admin123")
		assert.NoError(t, err)
		assert.True(t, check, "password must match")
		hpass = user.Password
		version = user.Version
	})

	t.Run("Update roles Admin", func(t *testing.T) {
		user := model.User{
			ID:      1,
			Version: version,
			Roles:   []model.Role{{ID: 1}, {ID: 3}},
		}
		err := store.User().Update(ctx, user, []model.UserField{model.UserField_Roles})
		assert.NoError(t, err)
	})

	t.Run("Get user Admin after update roles", func(t *testing.T) {
		user, err := store.User().Get(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)

		assert.Equal(t, 2, len(user.Roles))
		assert.Equal(t, "Admin", user.Roles[0].Name)
		assert.Equal(t, int64(3), user.Roles[1].ID)
		assert.Equal(t, "Staf", user.Roles[1].Name)

		version = user.Version
	})

	t.Run("Update password Admin", func(t *testing.T) {
		err := store.User().UpdatePassword(ctx, 1, version, "admin123")

		assert.NoError(t, err)
	})

	t.Run("Get user Admin", func(t *testing.T) {
		user, err := store.User().Get(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.NotEqual(t, "", user.Password)
		assert.NotEqual(t, "admin123", user.Password)

		check, err := user.VerifyPassword("admin123")
		assert.NoError(t, err)
		assert.True(t, check, "password must match")

		assert.NotEqual(t, hpass, user.Password, "hashed password must be different after update")
	})

	t.Run("Update invalid version user Admin", func(t *testing.T) {
		user := model.User{
			ID:        1,
			Version:   0,
			Email:     "admin@demo.com",
			UpdatedAt: updateTime,
		}

		err := store.User().Update(ctx, user, []model.UserField{model.UserField_Email, model.UserField_UpdatedAt})
		assert.ErrorContains(t, err, "no rows affected")
	})

	t.Run("Update user Admin", func(t *testing.T) {
		user := model.User{
			ID:        1,
			Version:   version,
			Email:     "admin@demo.com",
			UpdatedAt: updateTime,
		}

		err := store.User().Update(ctx, user, []model.UserField{model.UserField_Email, model.UserField_UpdatedAt})
		assert.NoError(t, err)
	})

	t.Run("Get user Admin after update", func(t *testing.T) {
		user, err := store.User().Get(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, int64(version+1), user.Version)
		assert.Equal(t, "admin@demo.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.Equal(t, createTime.Format(time.RFC3339), user.CreatedAt.Format(time.RFC3339), "createdAt must match")
		assert.Equal(t, updateTime.Format(time.RFC3339), user.UpdatedAt.Format(time.RFC3339), "updatedAt must match")
	})

	t.Run("Find all user", func(t *testing.T) {
		users, total, err := store.User().Find(ctx, nil, nil, 10, 0)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), total, "total must 1")
		assert.Equal(t, 1, len(users), "len(users) must 1")
		assert.Equal(t, "admin@demo.com", users[0].Email)
		assert.Equal(t, "Admin", users[0].Name)
	})

	t.Run("Create user Staff", func(t *testing.T) {
		user := model.User{
			Version:   1,
			Email:     "staff@demo.com",
			Name:      "Staff",
			CreatedAt: createTime1,
			CreatedBy: &model.User{
				ID: 1,
			},
			UpdatedAt: updateTime,
			UpdatedBy: &model.User{
				ID: 1,
			},
		}

		res, err := store.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(2), res.ID)
		assert.Equal(t, createTime1.Format(time.RFC3339), res.CreatedAt.Format(time.RFC3339), "createdAt must match")
		assert.Equal(t, updateTime.Format(time.RFC3339), res.UpdatedAt.Format(time.RFC3339), "updatedAt must match")
	})

	t.Run("Get user Staff", func(t *testing.T) {
		user, err := store.User().Get(ctx, 2)
		assert.NoError(t, err)

		assert.Equal(t, int64(2), user.ID)
		assert.Equal(t, "staff@demo.com", user.Email)
		assert.Equal(t, "Staff", user.Name)
		assert.NotNil(t, user.CreatedBy)
		assert.Equal(t, int64(1), user.CreatedBy.ID)
		assert.Equal(t, "Admin", user.CreatedBy.Name)
		assert.Equal(t, "admin@demo.com", user.CreatedBy.Email)
		assert.NotNil(t, user.UpdatedBy)
		assert.Equal(t, int64(1), user.UpdatedBy.ID)
		assert.Equal(t, "Admin", user.UpdatedBy.Name)
		assert.Equal(t, "admin@demo.com", user.UpdatedBy.Email)
	})

	t.Run("Create user Operator 1", func(t *testing.T) {
		user := model.User{
			Version: 1,
			Email:   "opr1@demo.com",
			Name:    "Operator 1",
		}

		res, err := store.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(3), res.ID)
	})

	t.Run("Create user Operator 2", func(t *testing.T) {
		user := model.User{
			Version: 1,
			Email:   "opr2@demo.com",
			Name:    "Operator 2",
		}

		res, err := store.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(4), res.ID)
	})

	t.Run("Create 23 dummy user", func(t *testing.T) {
		for i := 1; i <= 23; i++ {
			user := model.User{
				Version: 1,
				Email:   fmt.Sprintf("dummy%d@demo.com", i),
				Name:    fmt.Sprintf("Dummy %d", i),
			}

			res, err := store.User().Create(ctx, user)
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, int64(4+i), res.ID)
		}
	})

	t.Run("Find all user", func(t *testing.T) {
		users, total, err := store.User().Find(ctx, nil, []model.UserSorting{{
			Field: model.UserField_ID,
			Dir:   model.SortDir_ASC,
		}}, 10, 0)
		assert.NoError(t, err)

		assert.Equal(t, int64(27), total, "total must match")
		assert.Equal(t, 10, len(users), "len(users) must match")

		for i, u := range []string{"admin@demo.com", "staff@demo.com", "opr1@demo.com", "opr2@demo.com", "dummy1@demo.com"} {
			assert.Equal(t, u, users[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("Find all user limit 5 offset 5", func(t *testing.T) {
		users, total, err := store.User().Find(ctx, nil, nil, 5, 5)
		assert.NoError(t, err)

		assert.Equal(t, int64(27), total, "total must match")
		assert.Equal(t, 5, len(users), "len(users) must match")
		for i, u := range []string{"dummy2@demo.com", "dummy3@demo.com", "dummy4@demo.com"} {
			assert.Equal(t, u, users[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("Find users like dummy% limit 5", func(t *testing.T) {
		bv, _ := json.Marshal("dummy%")
		users, total, err := store.User().Find(ctx, []model.UserFilter{{
			Field: model.UserField_Email,
			Op:    model.FilterOp_Like,
			Value: json.RawMessage(bv),
		}}, nil, 5, 0)
		assert.NoError(t, err)

		assert.Equal(t, int64(23), total, "total must match")
		assert.Equal(t, 5, len(users), "len(users) must match")
		for i, u := range []string{"dummy1@demo.com", "dummy2@demo.com", "dummy3@demo.com", "dummy4@demo.com"} {
			assert.Equal(t, u, users[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("FindOne users like dummy5% sort by name asc", func(t *testing.T) {
		bv, _ := json.Marshal("dummy5%")
		user, err := store.User().FindOne(ctx, []model.UserFilter{{
			Field: model.UserField_Email,
			Op:    model.FilterOp_Like,
			Value: json.RawMessage(bv),
		}}, []model.UserSorting{{
			Field: model.UserField_Name,
			Dir:   model.SortDir_DESC,
		}})
		assert.NoError(t, err)
		assert.NotNil(t, user)

		assert.Equal(t, int64(9), user.ID)
		assert.Equal(t, "dummy5@demo.com", user.Email)
		assert.Equal(t, "Dummy 5", user.Name)
	})

	t.Run("Find users like dummy% sort by name desc limit 5", func(t *testing.T) {
		bv, _ := json.Marshal("dummy%")
		users, total, err := store.User().Find(ctx, []model.UserFilter{{
			Field: model.UserField_Email,
			Op:    model.FilterOp_Like,
			Value: json.RawMessage(bv),
		}}, []model.UserSorting{{
			Field: model.UserField_Name,
			Dir:   model.SortDir_DESC,
		}}, 5, 0)
		assert.NoError(t, err)

		assert.Equal(t, int64(23), total, "total must match")
		assert.Equal(t, 5, len(users), "len(users) must match")
		for i, u := range []string{"dummy9@demo.com", "dummy8@demo.com", "dummy7@demo.com", "dummy6@demo.com"} {
			assert.Equal(t, u, users[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})
}

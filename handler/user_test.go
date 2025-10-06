package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hanoman.co.id/crudgen/handler"
	"hanoman.co.id/crudgen/model"
	"hanoman.co.id/crudgen/util"
)

func TestUserCrudApi(t *testing.T) {
	logger := slog.New(util.NewPrettyHandler(&slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	store := model.GetStore()
	ctx := context.Background()

	api := http.NewServeMux()

	handler.UserHandlerRegister(api, store, handler.BasicAuthenticate)
	handler.RoleHandlerRegister(api, store, handler.BasicAuthenticate)
	handler.AuthHandlerRegister(api, store)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", handler.Secure(store, api)))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	t.Run("Test JWT", func(t *testing.T) {
		var err error
		token, err := handler.SignHS256(1, "admin@example.com", 2*time.Hour)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		_, err = handler.ParseHS256(token)
		assert.NoError(t, err)
	})

	t.Run("Create role Admin", func(t *testing.T) {
		role := model.Role{
			Name: "Admin",
			Privileges: `{
				"app_role":{"read":true,"create":true,"delete":true,"update":true},
				"app_user":{"read":true,"create":true,"delete":true,"update":true},
				"audit":{"read":true},
				"param":{"read":true,"create":true,"delete":true,"update":true},
				"identity":{"read":true,"create":true,"delete":true,"update":true},
				"tcpgwAccess":{"read":true,"create":true,"delete":true,"update":true},
				"tcpgwAccessIP":{"read":true,"create":true,"delete":true,"update":true},
				"kaltimtara":{"read":true,"create":false,"delete":false,"update":true}
			}`,
		}
		_, err := store.Role().Create(ctx, role)
		assert.NoError(t, err)
	})

	t.Run("Create user Admin", func(t *testing.T) {
		user := model.User{
			Version:  1,
			Email:    "admin@example.com",
			Name:     "Admin",
			Password: "admin123",
			Roles:    []model.Role{{ID: 1}},
		}

		res, err := store.User().Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.ID)
	})

	t.Run("Get user Admin before login", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/User/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Login with invalid salt", func(t *testing.T) {
		lreq := handler.LoginObject{
			Time:     time.Now().Format(time.RFC3339),
			Email:    "admin@example.com",
			Salt:     "Invalid SALT",
			Password: "admin123",
		}
		body, _ := json.Marshal(lreq)

		req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/auth", bytes.NewReader(body))
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	var token string
	var refreshToken string
	t.Run("Login", func(t *testing.T) {
		lreq := handler.LoginObject{
			Time:  time.Now().Format(time.RFC3339),
			Email: "admin@example.com",
		}
		body, _ := json.Marshal(lreq)

		req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/auth", bytes.NewReader(body))
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Invalid response: %s", string(bb)))
		assert.NotNil(t, resp.Body)

		err = json.Unmarshal(bb, &lreq)
		assert.NoError(t, err, fmt.Sprintf("Invalid response: %s", string(bb)))

		lreq.Password = "admin123"
		body, _ = json.Marshal(lreq)

		req, err = http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/auth", bytes.NewReader(body))
		assert.NoError(t, err)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bb, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Invalid response: %s", string(bb)))
		assert.NotNil(t, resp.Body)

		err = json.Unmarshal(bb, &lreq)
		assert.NoError(t, err, fmt.Sprintf("Invalid response: %s", string(bb)))

		assert.NotEmpty(t, lreq.Token, "token must not be empty")
		token = lreq.Token
		refreshToken = lreq.RefreshToken
	})

	t.Run("Get user Admin after login", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/User/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var user model.User
		err = json.Unmarshal(bb, &user)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)

		assert.NotEqual(t, "", user.Password)
		assert.NotEqual(t, "admin123", user.Password)

		check, err := user.VerifyPassword("admin123")
		assert.NoError(t, err)
		assert.True(t, check, "password must match")

		assert.Nil(t, user.CreatedBy)
		assert.Nil(t, user.UpdatedBy)
	})

	t.Run("RefreshToken", func(t *testing.T) {
		time.Sleep(5 * time.Second) // make sure the new token is different
		lreq := handler.LoginObject{
			Time:  time.Now().Format(time.RFC3339),
			Email: "admin@example.com",
		}
		body, _ := json.Marshal(lreq)

		req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/auth", bytes.NewReader(body))
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Invalid response: %s", string(bb)))
		assert.NotNil(t, resp.Body)

		err = json.Unmarshal(bb, &lreq)
		assert.NoError(t, err, fmt.Sprintf("Invalid response: %s", string(bb)))

		lreq.RefreshToken = refreshToken
		body, _ = json.Marshal(lreq)

		req, err = http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/auth", bytes.NewReader(body))
		assert.NoError(t, err)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bb, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Invalid response: %s", string(bb)))
		assert.NotNil(t, resp.Body)

		err = json.Unmarshal(bb, &lreq)
		assert.NoError(t, err, fmt.Sprintf("Invalid response: %s", string(bb)))

		assert.NotEmpty(t, lreq.Token, "token must not be empty")
		assert.NotEqual(t, token, lreq.Token, "token must be different after refresh")
		token = lreq.Token
		refreshToken = lreq.RefreshToken
	})

	t.Run("Get user Admin after RefreshToken", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/User/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var user model.User
		err = json.Unmarshal(bb, &user)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)

		assert.NotEqual(t, "", user.Password)
		assert.NotEqual(t, "admin123", user.Password)

		check, err := user.VerifyPassword("admin123")
		assert.NoError(t, err)
		assert.True(t, check, "password must match")

		assert.Nil(t, user.CreatedBy)
		assert.Nil(t, user.UpdatedBy)
	})

	t.Run("Create role Opr", func(t *testing.T) {
		role := model.Role{
			Name: "Opr",
			Privileges: `{
				"app_role":{"read":true},
				"app_user":{"read":true},
				"audit":{"read":true},
				"param":{"read":true},
				"identity":{"read":true},
				"tcpgwAccess":{"read":true},
				"tcpgwAccessIP":{"read":true,"create":true,"delete":true,"update":true},
				"kaltimtara":{"read":true,"create":false,"delete":false,"update":true}
			}`,
		}
		body, _ := json.Marshal(role)

		req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/Role", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(bb, &role)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), role.ID)
	})

	t.Run("Create role Staff", func(t *testing.T) {
		role := model.Role{
			Name: "Staf",
			Privileges: `{
				"app_role":{"read":true},
				"app_user":{"read":true},
				"audit":{"read":true},
				"param":{"read":true},
				"identity":{"read":true},
				"tcpgwAccess":{"read":true},
				"tcpgwAccessIP":{"read":true,"create":true,"delete":true,"update":true},
				"kaltimtara":{"read":true,"create":false,"delete":false,"update":true}
			}`,
		}
		body, _ := json.Marshal(role)

		req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/Role", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(bb, &role)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), role.ID)
	})

	var version int64
	t.Run("Get user Admin", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/User/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var user model.User
		err = json.Unmarshal(bb, &user)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, "Admin", user.Name)

		assert.NotEqual(t, "", user.Password)
		assert.NotEqual(t, "admin123", user.Password)

		check, err := user.VerifyPassword("admin123")
		assert.NoError(t, err)
		assert.True(t, check, "password must match")

		assert.Nil(t, user.CreatedBy)
		assert.Nil(t, user.UpdatedBy)
		version = user.Version
	})

	t.Run("Update user Admin", func(t *testing.T) {
		user := model.User{
			Version: version,
			Email:   "admin@demo.com",
			Roles:   []model.Role{{ID: 1}, {ID: 2}},
		}
		body, _ := json.Marshal(struct {
			Value  model.User `json:"value"`
			Fields []string   `json:"fields"`
		}{
			Value:  user,
			Fields: []string{"email", "roles"},
		})

		req, err := http.NewRequest(http.MethodPatch, "http://localhost:8080/api/v1/User/1", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(bb, &user)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)

		assert.Equal(t, "admin@demo.com", user.Email)
	})

	t.Run("Get user Admin after update with old token", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/User/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Login with new email", func(t *testing.T) {
		lreq := handler.LoginObject{
			Time:  time.Now().Format(time.RFC3339),
			Email: "admin@demo.com",
		}
		body, _ := json.Marshal(lreq)

		req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/auth", bytes.NewReader(body))
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Invalid response: %s", string(bb)))
		assert.NotNil(t, resp.Body)

		err = json.Unmarshal(bb, &lreq)
		assert.NoError(t, err, fmt.Sprintf("Invalid response: %s", string(bb)))

		lreq.Password = "admin123"
		body, _ = json.Marshal(lreq)

		req, err = http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/auth", bytes.NewReader(body))
		assert.NoError(t, err)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bb, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Invalid response: %s", string(bb)))
		assert.NotNil(t, resp.Body)

		err = json.Unmarshal(bb, &lreq)
		assert.NoError(t, err, fmt.Sprintf("Invalid response: %s", string(bb)))

		assert.NotEmpty(t, lreq.Token, "token must not be empty")
		token = lreq.Token
		refreshToken = lreq.RefreshToken
	})

	t.Run("Get user Admin after update", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/User/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var user model.User
		err = json.Unmarshal(bb, &user)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin@demo.com", user.Email)
		assert.Equal(t, "Admin", user.Name)
		assert.Equal(t, 2, len(user.Roles))
		assert.Equal(t, "Admin", user.Roles[0].Name)
		assert.Equal(t, "Opr", user.Roles[1].Name)
		assert.Nil(t, user.CreatedBy)
		assert.NotNil(t, user.UpdatedBy)
		assert.Equal(t, int64(1), user.UpdatedBy.ID)
	})

	t.Run("Create user Staff", func(t *testing.T) {
		user := model.User{
			Version: 1,
			Email:   "staff@demo.com",
			Name:    "Staff",
		}
		body, _ := json.Marshal(user)

		req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/User", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("response: %s", string(bb)))

		err = json.Unmarshal(bb, &user)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), user.ID)
	})

	t.Run("Get user Staff", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/User/2", nil)
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var user model.User
		err = json.Unmarshal(bb, &user)

		assert.Equal(t, int64(2), user.ID)
		assert.Equal(t, "staff@demo.com", user.Email)
		assert.Equal(t, "Staff", user.Name)
		assert.NotNil(t, user.CreatedBy)
		assert.Equal(t, int64(1), user.CreatedBy.ID)
		assert.NotNil(t, user.UpdatedBy)
		assert.Equal(t, int64(1), user.UpdatedBy.ID)
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
		body, _ := json.Marshal(struct {
			Limit   int                 `json:"limit"`
			Offset  int64               `json:"offset,omitempty"`
			Filter  []model.UserFilter  `json:"filter,omitempty"`
			Sorting []model.UserSorting `json:"sorting,omitempty"`
		}{
			Sorting: []model.UserSorting{{
				Field: model.UserField_ID,
				Dir:   model.SortDir_ASC,
			}},
			Limit: 10,
		})

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/User", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var result struct {
			List  []model.User `json:"list"`
			Total int64        `json:"total"`
		}
		err = json.Unmarshal(bb, &result)
		assert.NoError(t, err)

		assert.Equal(t, int64(27), result.Total)
		assert.Equal(t, 10, len(result.List))

		for i, u := range []string{"admin@demo.com", "staff@demo.com", "opr1@demo.com", "opr2@demo.com", "dummy1@demo.com"} {
			assert.Equal(t, u, result.List[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("Find all user limit 5 offset 5", func(t *testing.T) {
		body, _ := json.Marshal(struct {
			Limit   int                 `json:"limit"`
			Offset  int64               `json:"offset,omitempty"`
			Filter  []model.UserFilter  `json:"filter,omitempty"`
			Sorting []model.UserSorting `json:"sorting,omitempty"`
		}{
			Limit:  5,
			Offset: 5,
		})

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/User", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var result struct {
			List  []model.User `json:"list"`
			Total int64        `json:"total"`
		}
		err = json.Unmarshal(bb, &result)
		assert.NoError(t, err)

		assert.Equal(t, int64(27), result.Total)
		assert.Equal(t, 5, len(result.List), "len(users) must match")

		for i, u := range []string{"dummy2@demo.com", "dummy3@demo.com", "dummy4@demo.com"} {
			assert.Equal(t, u, result.List[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("Find users like dummy% limit 5", func(t *testing.T) {
		bv, _ := json.Marshal("dummy%")
		body, _ := json.Marshal(struct {
			Limit   int                 `json:"limit"`
			Offset  int64               `json:"offset,omitempty"`
			Filter  []model.UserFilter  `json:"filter,omitempty"`
			Sorting []model.UserSorting `json:"sorting,omitempty"`
		}{
			Limit:  5,
			Offset: 0,
			Filter: []model.UserFilter{{
				Field: model.UserField_Email,
				Op:    model.FilterOp_Like,
				Value: json.RawMessage(bv),
			}},
		})

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/User", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var result struct {
			List  []model.User `json:"list"`
			Total int64        `json:"total"`
		}
		err = json.Unmarshal(bb, &result)
		assert.NoError(t, err)

		assert.Equal(t, int64(23), result.Total, "total must match")
		assert.Equal(t, 5, len(result.List), "len(users) must match")
		for i, u := range []string{"dummy1@demo.com", "dummy2@demo.com", "dummy3@demo.com", "dummy4@demo.com"} {
			assert.Equal(t, u, result.List[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})

	t.Run("Find users like dummy% sort by name desc limit 5", func(t *testing.T) {
		bv, _ := json.Marshal("dummy%")
		body, _ := json.Marshal(struct {
			Limit   int                 `json:"limit"`
			Offset  int64               `json:"offset,omitempty"`
			Filter  []model.UserFilter  `json:"filter,omitempty"`
			Sorting []model.UserSorting `json:"sorting,omitempty"`
		}{
			Limit:  5,
			Offset: 0,
			Filter: []model.UserFilter{{
				Field: model.UserField_Email,
				Op:    model.FilterOp_Like,
				Value: json.RawMessage(bv),
			}},
			Sorting: []model.UserSorting{{
				Field: model.UserField_Name,
				Dir:   model.SortDir_DESC,
			}},
		})

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/User", bytes.NewReader(body))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, resp.Body)
		bb, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var result struct {
			List  []model.User `json:"list"`
			Total int64        `json:"total"`
		}
		err = json.Unmarshal(bb, &result)
		assert.NoError(t, err)

		assert.Equal(t, int64(23), result.Total, "total must match")
		assert.Equal(t, 5, len(result.List), "len(users) must match")
		for i, u := range []string{"dummy9@demo.com", "dummy8@demo.com", "dummy7@demo.com", "dummy6@demo.com"} {
			assert.Equal(t, u, result.List[i].Email, fmt.Sprintf("email must match at index %d", i))
		}
	})
}

package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"hanoman.co.id/crudgen/model"
)

func AuthHandlerRegister(mux *http.ServeMux, store model.Store) {
	mux.HandleFunc("PUT /auth", func(w http.ResponseWriter, r *http.Request) {
		if err := AuthLogin(r.Context(), store, w, r); err != nil {
			writeError(w, http.StatusInternalServerError, err)
		}
	})
	mux.HandleFunc("POST /auth", func(w http.ResponseWriter, r *http.Request) {
		if err := AuthRefresh(r.Context(), store, w, r); err != nil {
			writeError(w, http.StatusInternalServerError, err)
		}
	})
	mux.HandleFunc("DELETE /auth", func(w http.ResponseWriter, r *http.Request) {
		if err := AuthLogout(r.Context(), store, w, r); err != nil {
			writeError(w, http.StatusInternalServerError, err)
		}
	})
	mux.HandleFunc("GET /auth", func(w http.ResponseWriter, r *http.Request) {
		if err := AuthGet(r.Context(), store, w, r); err != nil {
			writeError(w, http.StatusInternalServerError, err)
		}
	})
	mux.HandleFunc("GET /auth/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
}

type LoginObject struct {
	Time         string     `json:"time"`
	Email        string     `json:"email"`
	Salt         string     `json:"salt,omitempty"`
	Password     string     `json:"password,omitempty"`
	Token        string     `json:"token,omitempty"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	User         *LoginUser `json:"user,omitempty"`
}

var (
	secret      = os.Getenv("JWT_SECRET")
	auth_secret = []byte(os.Getenv("AUTH_SECRET"))
)

func getUser(ctx context.Context, store model.Store, obj *LoginObject) *model.User {
	tt, err := time.Parse(time.RFC3339, obj.Time)
	if err != nil {
		slog.Warn("invalid time format", "err", err, "time", obj.Time)
		return nil
	}
	if tt.Before(time.Now().Add(-5*time.Minute)) || tt.After(time.Now().Add(5*time.Minute)) {
		slog.Warn("request time out of range", "time", obj.Time)
		return nil
	}
	if obj.Email == "" {
		slog.Warn("email is empty")
		return nil
	}
	h := hmac.New(sha256.New, auth_secret)
	h.Write([]byte(fmt.Sprintf("%s;%s", obj.Time, obj.Email)))
	if subtle.ConstantTimeCompare([]byte(obj.Salt), []byte(base64.RawStdEncoding.EncodeToString(h.Sum(nil)))) != 1 {
		slog.Warn("invalid salt", "email", obj.Email)
		return nil
	}
	user, err := store.User().GetByEmail(ctx, obj.Email)
	if err != nil {
		slog.Warn("failed to get user by email", "email", obj.Email, "err", err)
		return nil
	}
	if user == nil {
		slog.Warn("user not found", "email", obj.Email)
		return nil
	}
	if obj.Token != "" {
		claim, err := ParseHS256(obj.Token)
		if err != nil {
			slog.Warn("invalid token", "email", obj.Email, "err", err)
			return nil
		}
		if claim.UserID != user.ID {
			slog.Warn("user id mismatch", "email", obj.Email, "token_user_id", claim.UserID, "db_user_id", user.ID)
			return nil
		}
		if claim.Subject != user.Email {
			slog.Warn("email mismatch", "email", obj.Email, "token_email", claim.Subject)
			return nil
		}
		return user
	}
	if obj.RefreshToken != "" {
		claim, err := ParseHS256(obj.RefreshToken)
		if err != nil {
			slog.Warn("invalid token", "email", obj.Email, "err", err)
			return nil
		}
		if claim.UserID != user.ID {
			slog.Warn("user id mismatch", "email", obj.Email, "token_user_id", claim.UserID, "db_user_id", user.ID)
			return nil
		}
		if claim.Subject != user.Email {
			slog.Warn("email mismatch", "email", obj.Email, "token_email", claim.Subject)
			return nil
		}
		if user.Token.String != obj.RefreshToken {
			slog.Warn("refresh token mismatch", "email", obj.Email)
			return nil
		}
		return user
	}
	if obj.Password != "" {
		if obj.Salt == "" {
			slog.Warn("salt is empty", "email", obj.Email)
			return nil
		}
		h := hmac.New(sha256.New, auth_secret)
		h.Write([]byte(fmt.Sprintf("%s;%s", obj.Time, obj.Email)))
		if subtle.ConstantTimeCompare([]byte(obj.Salt), []byte(base64.RawStdEncoding.EncodeToString(h.Sum(nil)))) != 1 {
			slog.Warn("invalid salt", "email", obj.Email)
			return nil
		}
		if ok, err := VerifyPassword(obj.Password, user.Password); err != nil {
			slog.Warn("failed to verify password", "email", obj.Email, "err", err)
			return nil
		} else if !ok {
			slog.Warn("invalid password", "email", obj.Email)
			return nil
		}
		return user
	}
	return nil
}

func mergeMap(target map[string]any, source map[string]any) map[string]any {
	for k, sv := range source {
		if tv, ok := target[k]; ok {
			if svm, ok := sv.(map[string]any); ok {
				if tvm, ok := tv.(map[string]any); ok {
					target[k] = mergeMap(tvm, svm)
				} else {
					target[k] = sv
				}
			} else if svb, ok := sv.(bool); ok {
				if svb {
					target[k] = sv
				}
			} else {
				target[k] = sv
			}
		} else {
			target[k] = sv
		}
	}
	return target
}

func toLoginUser(user *model.User) *LoginUser {
	mprivs := map[string]any{}
	for _, role := range user.Roles {
		p := map[string]any{}
		err := json.Unmarshal([]byte(role.Privileges), &p)
		if err != nil {
			slog.Warn("failed to unmarshal role privileges", "role", role.Name, "privileges", role.Privileges, "err", err)
		}
		mprivs = mergeMap(mprivs, p)
	}
	return &LoginUser{
		Name:       user.Name,
		Email:      user.Email,
		Privileges: mprivs,
		User:       user,
	}
}

func AuthLogin(ctx context.Context, store model.Store, w http.ResponseWriter, r *http.Request) error {
	var obj LoginObject
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		slog.Warn("invalid body", "err", err)
		return fmt.Errorf("invalid body")
	}
	if obj.Salt == "" {
		h := hmac.New(sha256.New, auth_secret)
		h.Write([]byte(fmt.Sprintf("%s;%s", obj.Time, obj.Email)))
		obj.Salt = base64.RawStdEncoding.EncodeToString(h.Sum(nil))
		obj.Token = ""
		obj.RefreshToken = ""
		obj.Password = ""
		_ = json.NewEncoder(w).Encode(obj)
		return nil
	}
	obj.Token = ""
	obj.RefreshToken = ""
	user := getUser(ctx, store, &obj)
	obj.Password = ""
	if user == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	obj.Token, err = SignHS256(user.ID, user.Email, 15*time.Minute)
	if err != nil {
		slog.Error("failed to sign token", "err", err)
		return fmt.Errorf("login failed")
	}
	obj.RefreshToken, err = SignHS256(user.ID, user.Email, 1*time.Hour)
	if err != nil {
		slog.Error("failed to sign refresh token", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	user.Token = sql.NullString{String: obj.RefreshToken, Valid: true}
	err = store.User().Update(ctx, *user, []model.UserField{model.UserField_Token})
	if err != nil {
		slog.Error("failed to update user token", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	obj.User = toLoginUser(user)
	_ = json.NewEncoder(w).Encode(obj)
	return nil
}

func AuthRefresh(ctx context.Context, store model.Store, w http.ResponseWriter, r *http.Request) error {
	var obj LoginObject
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		slog.Warn("invalid body", "err", err)
		return fmt.Errorf("invalid body")
	}
	obj.Password = ""
	obj.Token = ""
	user := getUser(ctx, store, &obj)
	obj.RefreshToken = ""
	if user == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	obj.Token, err = SignHS256(user.ID, user.Email, 15*time.Minute)
	if err != nil {
		slog.Error("failed to sign token", "err", err)
		return fmt.Errorf("login failed")
	}
	obj.RefreshToken, err = SignHS256(user.ID, user.Email, 1*time.Hour)
	if err != nil {
		slog.Error("failed to sign refresh token", "err", err)
		return fmt.Errorf("login failed")
	}
	user.Token = sql.NullString{String: obj.RefreshToken, Valid: true}
	err = store.User().Update(ctx, *user, []model.UserField{model.UserField_Token})
	if err != nil {
		slog.Error("failed to update user token", "err", err)
		return fmt.Errorf("login failed")
	}
	obj.User = toLoginUser(user)
	_ = json.NewEncoder(w).Encode(obj)
	return nil
}

func AuthLogout(ctx context.Context, store model.Store, w http.ResponseWriter, r *http.Request) error {
	var obj LoginObject
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		slog.Warn("invalid body", "err", err)
		return fmt.Errorf("invalid body")
	}
	obj.Password = ""
	obj.RefreshToken = ""
	user := getUser(ctx, store, &obj)
	obj.Token = ""
	if user == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	user.Token = sql.NullString{String: "", Valid: false}
	err = store.User().Update(ctx, *user, []model.UserField{model.UserField_Token})
	if err != nil {
		slog.Warn("failed to update user token", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
	return nil
}

func AuthGet(ctx context.Context, store model.Store, w http.ResponseWriter, r *http.Request) error {
	var obj LoginObject
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		slog.Warn("invalid body", "err", err)
		return fmt.Errorf("invalid body")
	}
	obj.Password = ""
	obj.RefreshToken = ""
	user := getUser(ctx, store, &obj)
	obj.Token = ""
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		return nil
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Roles: user.Roles,
	})
	return nil
}

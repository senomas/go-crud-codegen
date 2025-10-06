package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"hanoman.co.id/crudgen/model"
)

type Params struct {
	Memory      uint32 // in KiB
	Iterations  uint32
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

var hashParam = &Params{
	Memory:      64 * 1024, // 64 MiB
	Iterations:  3,
	Parallelism: 2,
	SaltLen:     16,
	KeyLen:      32,
}

type LoginUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`

	User       *model.User
	Privileges map[string]any
}

type Authenticate func(r *http.Request, resourece, action string) bool

func Secure(store model.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var claim *JwtClaims
		var err error
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := auth[7:]
			claim, err = ParseHS256(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("unauthorized"))
				return
			}
		}
		if claim == nil {
			session, _ := r.Cookie("session")
			if session != nil {
				claim, err = ParseHS256(session.Value)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("unauthorized"))
					return
				}
			}
		}
		if claim != nil {
			user, err := store.User().Get(r.Context(), claim.UserID)
			if err != nil {
				slog.Error("get user by id", "user_id", claim.UserID, "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("unauthorized"))
				return
			}
			if user.Email != claim.Subject {
				slog.Warn("user email not match", "user_id", claim.UserID, "email", user.Email, "subject", claim.Subject)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("unauthorized"))
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(),
				HandlerCtxKeyUser, toLoginUser(user))))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func BasicAuthenticate(r *http.Request, resource, action string) bool {
	if luser, ok := r.Context().Value(HandlerCtxKeyUser).(*LoginUser); !ok {
		slog.Warn("missing login user in context")
	} else if luser == nil {
		slog.Warn("nil login user in context")
	} else if pm, ok := luser.Privileges[resource].(map[string]any); ok {
		if b, ok := pm[action].(bool); ok {
			return b
		}
		slog.Warn("missing privilege for action", "resource", resource, "action", action, "user", luser.Email, "res.privileges", pm)
	} else {
		slog.Warn("missing privilege for resource", "resource", resource, "user", luser.Email, "privileges", luser.Privileges)
	}
	return false
}

func HashPassword(value string) (string, error) {
	salt := make([]byte, hashParam.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(value),
		salt, hashParam.Iterations,
		hashParam.Memory,
		hashParam.Parallelism,
		hashParam.KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	phc := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		hashParam.Memory,
		hashParam.Iterations,
		hashParam.Parallelism,
		b64Salt,
		b64Hash)
	return phc, nil
}

func VerifyPassword(value, password string) (bool, error) {
	parts := strings.Split(password, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("invalid PHC format")
	}
	if parts[2] != "v=19" {
		return false, errors.New("unsupported argon2 version")
	}

	var mem, iters uint32
	var par uint8
	for _, kv := range strings.Split(parts[3], ",") {
		k, v, _ := strings.Cut(kv, "=")
		switch k {
		case "m":
			u, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return false, err
			}
			mem = uint32(u)
			if mem < 8*1024 {
				return false, errors.New("insufficient memory")
			}
		case "t":
			u, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return false, err
			}
			iters = uint32(u)
			if iters < 3 {
				return false, errors.New("insufficient iterations")
			}
		case "p":
			u, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return false, err
			}
			par = uint8(u)
			if par < 2 {
				return false, errors.New("insufficient parallelism")
			}
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	got := argon2.IDKey([]byte(value), salt, iters, mem, par, uint32(len(want)))
	ok := subtle.ConstantTimeCompare(got, want) == 1
	return ok, nil
}

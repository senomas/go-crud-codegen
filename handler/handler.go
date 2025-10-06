package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type HandlerCtxKey string

const (
	HandlerCtxKeyUser HandlerCtxKey = "user"
)

type HttpError struct {
	Error string `json:"error"`
}

type AccessPermission func(resource, action string) bool

var jwtSecret = os.Getenv("JWT_SECRET")

type JwtClaims struct {
	UserID     int64          `json:"uid"`
	Privileges map[string]any `json:"privileges,omitempty"`
	jwt.RegisteredClaims
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	merr := HttpError{Error: err.Error()}
	json.NewEncoder(w).Encode(merr)
}

func SignHS256(uid int64, subject string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JwtClaims{
		UserID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    "mwui",
			Audience:  []string{"mwui-clients"},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(jwtSecret))
}

func ParseHS256(tokenStr string) (*JwtClaims, error) {
	var claims JwtClaims
	tok, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected alg")
		}
		return []byte(jwtSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(30*time.Second),
		jwt.WithIssuedAt(),
		jwt.WithIssuer("mwui"),
		jwt.WithAudience("mwui-clients"),
	)
	if err != nil {
		slog.Debug("token parse error", "err", err)
		return nil, err
	}
	if !tok.Valid {
		slog.Debug("invalid token")
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

package model

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"
)

type FilterOp string

const (
	FilterOp_EQ    FilterOp = "eq"
	FilterOp_Like  FilterOp = "like"
	FilterOp_ILike FilterOp = "ilike"
)

type SortDir string

const (
	SortDir_ASC  SortDir = "asc"
	SortDir_DESC SortDir = "desc"
)

type Store interface {
	User() UserStore
	Role() RoleStore
}

type StoreImpl struct {
	db *sql.DB
}

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

func GetStore() Store {
	driver := "sqlite3"
	db, err := sql.Open(driver, "memory:")
	if err != nil {
		slog.Error("Error opening db", "error", err)
		os.Exit(1)
	}
	return &StoreImpl{db: db}
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
		case "t":
			u, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return false, err
			}
			iters = uint32(u)
		case "p":
			u, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return false, err
			}
			par = uint8(u)
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

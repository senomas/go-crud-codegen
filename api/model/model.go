package model

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
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

type Repos interface {
	User() UserRepository
}

type RepositoryImpl struct {
	db *sql.DB
}

func GetRepos() Repos {
	dsn := "file:../app.db?_busy_timeout=5000&cache=shared&mode=rwc&_foreign_keys=on"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) User() UserRepository {
	return &UserRepositoryImpl{
		RepositoryImpl: r,
	}
}

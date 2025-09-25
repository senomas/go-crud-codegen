package model

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

type Store interface {
	User() UserStore
	Role() RoleStore
}

type StoreImpl struct {
	db *sql.DB
}

func GetStore() Store {
	dsn := "file:../app.db?_busy_timeout=5000&cache=shared&mode=rwc&_foreign_keys=on"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	dir := "../migrations"
	ents, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	files := make([]string, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.EqualFold(filepath.Ext(name), ".sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(filepath.Base(files[i])) < strings.ToLower(filepath.Base(files[j]))
	})
	fmt.Printf("Migrate %+v files\n", files)
	for _, f := range files {
		err = RunSQLFile(context.Background(), db, f)
		if err != nil {
			log.Fatalf("migrate %s: %v", f, err)
		}
	}

	return &StoreImpl{db: db}
}

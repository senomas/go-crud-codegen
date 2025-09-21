package model

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

func NewUserRepository(ctx context.Context, db *sql.DB) *UserRepositoryImpl {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS app_user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR(255) UNIQUE,
		name VARCHAR(255),
		salt VARCHAR(255),
		password VARCHAR(255),
		token VARCHAR(255)
	)`)
	if err != nil {
		log.Panic(err)
	}
	total := 0
	err = db.QueryRowContext(ctx, "\nSELECT COUNT(*) FROM \napp_user WHERE name LIKE ?", "%foo%").Scan(&total)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Total users with name like:", total)
	return &UserRepositoryImpl{
		ctx: ctx,
		db:  db,
	}
}

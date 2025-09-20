package model

import (
	"context"
	"database/sql"
	"log"
)

type StringFilterOp string

const (
	StringFilterOp_EQ    StringFilterOp = "eq"
	StringFilterOp_Like  StringFilterOp = "like"
	StringFilterOp_ILike StringFilterOp = "ilike"
)

type StringFilter struct {
	Op    StringFilterOp `json:"op"`
	Value *string        `json:"value"`
}

type IntFilterOp string

const (
	IntFilterOp_EQ      StringFilterOp = "eq"
	IntFilterOp_Between StringFilterOp = "between"
)

type IntFilter struct {
	Op     IntFilterOp `json:"op"`
	Value  *int64      `json:"value"`
	Value2 *int64      `json:"value2"`
}

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
	return &UserRepositoryImpl{
		ctx: ctx,
		db:  db,
	}
}

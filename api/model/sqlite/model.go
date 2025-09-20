package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"hanoman.co.id/mwui/api/model"
)

type repos struct {
	ctx context.Context
	tx  *sql.Tx
}

func init() {
	dir, _ := os.Getwd()
	dsn := fmt.Sprintf("file:%s/../../app.db?_busy_timeout=5000&cache=shared&mode=rwc&?_foreign_keys=on", dir)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)

	_, err = db.ExecContext(context.Background(), `
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS app_user(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			name  TEXT NOT NULL,
			password TEXT NOT NULL,
			salt TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			created_by INTEGER NOT NULL,
			updated_at DATETIME,
			updated_by INTEGER
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	model.Register("sqlite", func(ctx context.Context) (model.Repos, context.Context, error) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, ctx, err
		}
		return &repos{
			ctx: ctx,
			tx:  tx,
		}, ctx, nil
	})
}

func (r *repos) User() model.UserRepository {
	return &UserRepositoryImpl{repos: r}
}

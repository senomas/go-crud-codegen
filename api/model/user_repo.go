package model

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type UserRepositoryImpl struct {
	ctx context.Context
	tx  *sql.Tx
	db  *goqu.Database
}

func (r *UserRepositoryImpl) Create(user User) (int64, error) {
	defer r.tx.Commit()
	sql, args, err := r.db.Insert("app_user").Rows(user).ToSQL()
	if err != nil {
		return 0, err
	}
	log.Printf("Create.user %v\n%v\n", sql, args)
	res, err := r.tx.ExecContext(r.ctx, `
		INSERT INTO app_user(
			email,
			name,
			salt,
			password,
			created_by,
			created_at,
			updated_by,
			updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $5, $6)
	`, user.Email, user.Name, user.Salt, user.Password, user.CreatedBy.ID, time.Now())
	if err != nil {
		r.tx.Rollback()
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepositoryImpl) Update(user User) error {
	return nil
}

func (r *UserRepositoryImpl) Delete(id int64) error {
	return nil
}

func (r *UserRepositoryImpl) Get(id int64) (*User, error) {
	defer r.tx.Commit()
	row := r.tx.QueryRowContext(r.ctx, `
		SELECT
			id,
			email,
			name,
			salt,
			password,
			created_by,
			created_at,
			updated_by,
			updated_at
		FROM
			app_user
		WHERE id = $1
	`, id)
	var user User
	var createdBy int64
	var createdAt time.Time
	var updatedBy int64
	var updatedAt time.Time
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Salt,
		&user.Password,
		&createdBy,
		&createdAt,
		&updatedBy,
		&updatedAt,
	)
	if err != nil {
		r.tx.Rollback()
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Find(filter UserFilter) ([]User, int64, error) {
	return nil, 0, nil
}

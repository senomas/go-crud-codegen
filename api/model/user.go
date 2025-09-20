package model

import (
	"time"
)

type Role struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UserRef struct {
	ID      int64  `json:"id"`
	Version int64  `json:"version"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

type User struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Salt      string     `json:"salt"`
	Password  string     `json:"password"`
	Token     *string    `json:"token,omitempty"`
	Roles     []Role     `json:"roles"`
	CreatedBy UserRef    `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedBy *UserRef   `json:"updated_by,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type UserSortFields string

const (
	User_ID    UserSortFields = "id"
	User_Email UserSortFields = "email"
	User_Name  UserSortFields = "name"
)

type UserSort struct {
	Field UserSortFields `json:"field"`
	Dir   SortDir        `json:"dir"`
}

type UserFilter struct {
	Email StringFilter `json:"email"`
	Name  StringFilter `json:"name"`
	Sorts []SortDir    `json:"sorts"`
}

type UserRepository interface {
	Create(user User) (int64, error)
	Update(user User) error
	Delete(id int64) error
	Get(id int64) (*User, error)
	Find(filter UserFilter) ([]User, int64, error)
}

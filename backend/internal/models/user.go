package models

import (
	"time"
)

// User represents an application user (admin or regular)
type User struct {
	UserID       int        `db:"user_id" json:"user_id"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         string     `db:"role" json:"role"`
	FirstName    string     `db:"first_name" json:"first_name"`
	LastName     string     `db:"last_name" json:"last_name"`
	Email        string     `db:"email" json:"email"`
	Phone        *string    `db:"phone" json:"phone,omitempty"`
	Department   *string    `db:"department" json:"department,omitempty"`
	Position     *string    `db:"position" json:"position,omitempty"`
	LastLogin    *time.Time `db:"last_login" json:"last_login,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

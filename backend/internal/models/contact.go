package models

import (
	"time"
)

// Contact represents an individual contact of a customer
type Contact struct {
	ContactID  int       `db:"contact_id" json:"contact_id"`
	CustomerID int       `db:"customer_id" json:"customer_id"`
	FirstName  string    `db:"first_name" json:"first_name"`
	LastName   string    `db:"last_name" json:"last_name"`
	Position   *string   `db:"position" json:"position,omitempty"`
	Phone      *string   `db:"phone" json:"phone,omitempty"`
	Email      *string   `db:"email" json:"email,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

package models

import (
	"time"
)

// Customer represents a client company
type Customer struct {
	CustomerID  int       `db:"customer_id" json:"customer_id"`
	CompanyName string    `db:"company_name" json:"company_name"`
	Industry    *string   `db:"industry" json:"industry,omitempty"`
	Address     *string   `db:"address" json:"address,omitempty"`
	Phone       *string   `db:"phone" json:"phone,omitempty"`
	Email       *string   `db:"email" json:"email,omitempty"`
	Website     *string   `db:"website" json:"website,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

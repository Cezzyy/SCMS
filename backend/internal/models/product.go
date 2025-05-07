package models

import (
	"encoding/json"
	"time"
)

// Product maintains equipment details
type Product struct {
	ProductID       int             `db:"product_id" json:"product_id"`
	ProductName     string          `db:"product_name" json:"product_name"`
	Model           *string         `db:"model" json:"model,omitempty"`
	Description     *string         `db:"description" json:"description,omitempty"`
	TechnicalSpecs  json.RawMessage `db:"technical_specs" json:"technical_specs,omitempty"`
	Certifications  *string         `db:"certifications" json:"certifications,omitempty"`
	SafetyStandards *string         `db:"safety_standards" json:"safety_standards,omitempty"`
	WarrantyPeriod  int             `db:"warranty_period" json:"warranty_period"`
	Price           float64         `db:"price" json:"price"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

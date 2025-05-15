package models

import (
	"time"
)

// Quotation stores generated quotes
type Quotation struct {
	QuotationID  int       `db:"quotation_id" json:"quotation_id"`
	CustomerID   int       `db:"customer_id" json:"customer_id"`
	QuoteDate    time.Time `db:"quote_date" json:"quote_date"`
	ValidityDate time.Time `db:"validity_date" json:"validity_date"`
	Status       string    `db:"status" json:"status"`
	TotalAmount  float64   `db:"total_amount" json:"total_amount"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// QuotationItem details each line in a quotation
type QuotationItem struct {
	QuotationItemID int     `db:"quotation_item_id" json:"quotation_item_id"`
	QuotationID     int     `db:"quotation_id" json:"quotation_id"`
	ProductID       int     `db:"product_id" json:"product_id"`
	Quantity        int     `db:"quantity" json:"quantity"`
	UnitPrice       float64 `db:"unit_price" json:"unit_price"`
	Discount        float64 `db:"discount" json:"discount"`
	LineTotal       float64 `db:"line_total" json:"line_total"`
}

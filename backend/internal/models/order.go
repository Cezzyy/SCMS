package models

import (
	"time"
)

// Order records sales transactions
type Order struct {
	OrderID         int       `db:"order_id" json:"order_id"`
	CustomerID      int       `db:"customer_id" json:"customer_id"`
	QuotationID     *int      `db:"quotation_id" json:"quotation_id,omitempty"`
	OrderDate       time.Time `db:"order_date" json:"order_date"`
	ShippingAddress string    `db:"shipping_address" json:"shipping_address"`
	Status          string    `db:"status" json:"status"`
	TotalAmount     float64   `db:"total_amount" json:"total_amount"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// OrderItem lists products within an order
type OrderItem struct {
	OrderItemID int     `db:"order_item_id" json:"order_item_id"`
	OrderID     int     `db:"order_id" json:"order_id"`
	ProductID   int     `db:"product_id" json:"product_id"`
	Quantity    int     `db:"quantity" json:"quantity"`
	UnitPrice   float64 `db:"unit_price" json:"unit_price"`
	Discount    float64 `db:"discount" json:"discount"`
	LineTotal   float64 `db:"line_total" json:"line_total"`
}

package models

import (
	"time"
)

// Inventory tracks stock levels
type Inventory struct {
	InventoryID     int        `db:"inventory_id" json:"inventory_id"`
	ProductID       int        `db:"product_id" json:"product_id"`
	CurrentStock    int        `db:"current_stock" json:"current_stock"`
	ReorderLevel    int        `db:"reorder_level" json:"reorder_level"`
	LastRestockDate *time.Time `db:"last_restock_date" json:"last_restock_date,omitempty"`
} 
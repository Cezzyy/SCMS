package models

import "time"

// SalesTrend represents daily sales data for the sales trends report
type SalesTrend struct {
	Day         string  `json:"day" db:"day"`
	TotalAmount float64 `json:"total_amount" db:"total_amount"`
}

// LowStockItem represents inventory items below reorder level
type LowStockItem struct {
	ID           int     `json:"id" db:"inventory_id"`
	ProductID    int     `json:"product_id" db:"product_id"`
	ProductName  string  `json:"name" db:"product_name"`
	CurrentStock int     `json:"current_stock" db:"current_stock"`
	ReorderLevel int     `json:"reorder_level" db:"reorder_level"`
	UnitPrice    float64 `json:"unit_price" db:"unit_price"`
}

// TopCustomer represents customer with highest sales values
type TopCustomer struct {
	ID          int     `json:"id" db:"customer_id"`
	Name        string  `json:"name" db:"company_name"`
	TotalSpent  float64 `json:"total_spent" db:"total_spent"`
	OrderCount  int     `json:"orders" db:"order_count"`
	ContactName string  `json:"contact_name,omitempty" db:"contact_name"`
}

// DashboardSummary represents the complete dashboard data
type DashboardSummary struct {
	TotalSales    float64        `json:"total_sales"`
	OrderCount    int            `json:"order_count"`
	LowStockCount int            `json:"low_stock_count"`
	SalesTrends   []SalesTrend   `json:"sales_trends"`
	LowStockItems []LowStockItem `json:"low_stock_items"`
	TopCustomers  []TopCustomer  `json:"top_customers"`
	Period        string         `json:"period"`
	LastUpdated   time.Time      `json:"last_updated"`
}

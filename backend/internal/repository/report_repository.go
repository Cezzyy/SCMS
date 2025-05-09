package repository

import (
	"context"
	"time"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/jmoiron/sqlx"
)

// ReportRepository handles database operations for reports and dashboard data
type ReportRepository struct {
	db *sqlx.DB
}

// NewReportRepository creates a new repository with the provided database connection
func NewReportRepository(db *sqlx.DB) *ReportRepository {
	return &ReportRepository{
		db: db,
	}
}

// GetSalesTrends retrieves sales data for the specified number of days
func (r *ReportRepository) GetSalesTrends(ctx context.Context, days int) ([]models.SalesTrend, error) {
	trends := []models.SalesTrend{}

	query := `
		SELECT 
			TO_CHAR(order_date, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(total_amount), 0) AS total_amount
		FROM 
			orders
		WHERE 
			order_date >= CURRENT_DATE - INTERVAL '$1 days'
		GROUP BY 
			day
		ORDER BY 
			day ASC
	`
	err := r.db.SelectContext(ctx, &trends, query, days)
	return trends, err
}

// GetTotalSales retrieves the total sales amount for the specified number of days
func (r *ReportRepository) GetTotalSales(ctx context.Context, days int) (float64, error) {
	var totalSales float64

	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) AS total_sales
		FROM 
			orders
		WHERE 
			order_date >= CURRENT_DATE - INTERVAL '$1 days'
	`
	err := r.db.GetContext(ctx, &totalSales, query, days)
	return totalSales, err
}

// GetOrderCount retrieves the total number of orders for the specified number of days
func (r *ReportRepository) GetOrderCount(ctx context.Context, days int) (int, error) {
	var orderCount int

	query := `
		SELECT 
			COUNT(*) AS order_count
		FROM 
			orders
		WHERE 
			order_date >= CURRENT_DATE - INTERVAL '$1 days'
	`
	err := r.db.GetContext(ctx, &orderCount, query, days)
	return orderCount, err
}

// GetLowStockItems retrieves inventory items that are below their reorder level
func (r *ReportRepository) GetLowStockItems(ctx context.Context) ([]models.LowStockItem, error) {
	items := []models.LowStockItem{}

	query := `
		SELECT 
			i.inventory_id,
			i.product_id,
			p.name AS product_name,
			i.current_stock,
			i.reorder_level,
			p.unit_price
		FROM 
			inventory i
		INNER JOIN 
			products p ON i.product_id = p.product_id
		WHERE 
			i.current_stock < i.reorder_level
		ORDER BY 
			(i.reorder_level - i.current_stock) DESC
	`
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

// GetLowStockCount retrieves the count of inventory items below reorder level
func (r *ReportRepository) GetLowStockCount(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT 
			COUNT(*) AS low_stock_count
		FROM 
			inventory
		WHERE 
			current_stock < reorder_level
	`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

// GetTopCustomers retrieves the top customers by total order amount
func (r *ReportRepository) GetTopCustomers(ctx context.Context, limit int, days int) ([]models.TopCustomer, error) {
	customers := []models.TopCustomer{}

	query := `
		SELECT 
			c.customer_id,
			c.company_name,
			COALESCE(SUM(o.total_amount), 0) AS total_spent,
			COUNT(o.order_id) AS order_count,
			(
				SELECT co.first_name || ' ' || co.last_name 
				FROM contacts co 
				WHERE co.customer_id = c.customer_id 
				LIMIT 1
			) AS contact_name
		FROM 
			customers c
		LEFT JOIN 
			orders o ON c.customer_id = o.customer_id AND o.order_date >= CURRENT_DATE - INTERVAL '$2 days'
		GROUP BY 
			c.customer_id
		ORDER BY 
			total_spent DESC
		LIMIT $1
	`
	err := r.db.SelectContext(ctx, &customers, query, limit, days)
	return customers, err
}

// GetDashboardSummary retrieves all dashboard data in a single request
func (r *ReportRepository) GetDashboardSummary(ctx context.Context, days int) (models.DashboardSummary, error) {
	var summary models.DashboardSummary
	var err error

	// Get sales trends
	summary.SalesTrends, err = r.GetSalesTrends(ctx, days)
	if err != nil {
		return summary, err
	}

	// Get total sales
	summary.TotalSales, err = r.GetTotalSales(ctx, days)
	if err != nil {
		return summary, err
	}

	// Get order count
	summary.OrderCount, err = r.GetOrderCount(ctx, days)
	if err != nil {
		return summary, err
	}

	// Get low stock items
	summary.LowStockItems, err = r.GetLowStockItems(ctx)
	if err != nil {
		return summary, err
	}

	// Get low stock count
	summary.LowStockCount = len(summary.LowStockItems)

	// Get top customers (limit to 5)
	summary.TopCustomers, err = r.GetTopCustomers(ctx, 5, days)
	if err != nil {
		return summary, err
	}

	// Set period and last updated
	summary.Period = "Last " + time.Now().AddDate(0, 0, -days).Format("Jan 2") + " - " + time.Now().Format("Jan 2")
	summary.LastUpdated = time.Now()

	return summary, nil
}

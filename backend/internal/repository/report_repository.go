package repository

import (
	"context"
	"fmt"
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

	fmt.Printf("Executing GetSalesTrends query with days=%d\n", days)

	query := `
		SELECT 
			TO_CHAR(order_date, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(total_amount), 0) AS total_amount
		FROM 
			orders
		WHERE 
			order_date >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY 
			day
		ORDER BY 
			day ASC
	`

	// Format the query with the days parameter directly
	formattedQuery := fmt.Sprintf(query, days)
	fmt.Printf("Formatted query: %s\n", formattedQuery)

	err := r.db.SelectContext(ctx, &trends, formattedQuery)
	if err != nil {
		fmt.Printf("Error executing sales trends query: %v\n", err)
		return trends, err
	}

	fmt.Printf("Retrieved %d sales trend records\n", len(trends))
	return trends, nil
}

// GetTotalSales retrieves the total sales amount for the specified number of days
func (r *ReportRepository) GetTotalSales(ctx context.Context, days int) (float64, error) {
	var totalSales float64

	fmt.Printf("Executing GetTotalSales query with days=%d\n", days)

	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) AS total_sales
		FROM 
			orders
		WHERE 
			order_date >= CURRENT_DATE - INTERVAL '%d days'
	`

	// Format the query with the days parameter directly
	formattedQuery := fmt.Sprintf(query, days)
	fmt.Printf("Formatted query: %s\n", formattedQuery)

	err := r.db.GetContext(ctx, &totalSales, formattedQuery)
	if err != nil {
		fmt.Printf("Error executing total sales query: %v\n", err)
		return totalSales, err
	}

	fmt.Printf("Total sales: %.2f\n", totalSales)
	return totalSales, nil
}

// GetOrderCount retrieves the total number of orders for the specified number of days
func (r *ReportRepository) GetOrderCount(ctx context.Context, days int) (int, error) {
	var orderCount int

	fmt.Printf("Executing GetOrderCount query with days=%d\n", days)

	query := `
		SELECT 
			COUNT(*) AS order_count
		FROM 
			orders
		WHERE 
			order_date >= CURRENT_DATE - INTERVAL '%d days'
	`

	// Format the query with the days parameter directly
	formattedQuery := fmt.Sprintf(query, days)
	fmt.Printf("Formatted query: %s\n", formattedQuery)

	err := r.db.GetContext(ctx, &orderCount, formattedQuery)
	if err != nil {
		fmt.Printf("Error executing order count query: %v\n", err)
		return orderCount, err
	}

	fmt.Printf("Order count: %d\n", orderCount)
	return orderCount, nil
}

// GetLowStockItems retrieves inventory items that are below their reorder level
func (r *ReportRepository) GetLowStockItems(ctx context.Context) ([]models.LowStockItem, error) {
	items := []models.LowStockItem{}

	fmt.Printf("Executing GetLowStockItems query\n")

	// Adjust the query to use price instead of unit_price which is the correct column name per the schema
	query := `
		SELECT 
			i.inventory_id,
			i.product_id,
			p.product_name AS product_name,
			i.current_stock,
			i.reorder_level,
			p.price AS unit_price
		FROM 
			inventory i
		INNER JOIN 
			products p ON i.product_id = p.product_id
		WHERE 
			i.current_stock < i.reorder_level
		ORDER BY 
			(i.reorder_level - i.current_stock) DESC
	`

	fmt.Printf("Query: %s\n", query)

	err := r.db.SelectContext(ctx, &items, query)
	if err != nil {
		fmt.Printf("Error executing low stock items query: %v\n", err)
		return items, err
	}

	fmt.Printf("Retrieved %d low stock items\n", len(items))
	return items, nil
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

	fmt.Printf("Executing GetTopCustomers query with limit=%d, days=%d\n", limit, days)

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
			orders o ON c.customer_id = o.customer_id AND o.order_date >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY 
			c.customer_id
		ORDER BY 
			total_spent DESC
		LIMIT %d
	`

	// Format the query with the days and limit parameters directly
	formattedQuery := fmt.Sprintf(query, days, limit)
	fmt.Printf("Formatted query: %s\n", formattedQuery)

	err := r.db.SelectContext(ctx, &customers, formattedQuery)
	if err != nil {
		fmt.Printf("Error executing top customers query: %v\n", err)
		return customers, err
	}

	fmt.Printf("Retrieved %d top customer records\n", len(customers))
	return customers, nil
}

// GetDashboardSummary retrieves all dashboard data in a single request
func (r *ReportRepository) GetDashboardSummary(ctx context.Context, days int) (models.DashboardSummary, error) {
	var summary models.DashboardSummary
	var err error

	fmt.Printf("Getting dashboard summary for past %d days\n", days)

	// Get sales trends
	summary.SalesTrends, err = r.GetSalesTrends(ctx, days)
	if err != nil {
		fmt.Printf("Error getting sales trends: %v\n", err)
		return summary, fmt.Errorf("error getting sales trends: %w", err)
	}

	// Get total sales
	summary.TotalSales, err = r.GetTotalSales(ctx, days)
	if err != nil {
		fmt.Printf("Error getting total sales: %v\n", err)
		return summary, fmt.Errorf("error getting total sales: %w", err)
	}

	// Get order count
	summary.OrderCount, err = r.GetOrderCount(ctx, days)
	if err != nil {
		fmt.Printf("Error getting order count: %v\n", err)
		return summary, fmt.Errorf("error getting order count: %w", err)
	}

	// Get low stock items
	summary.LowStockItems, err = r.GetLowStockItems(ctx)
	if err != nil {
		fmt.Printf("Error getting low stock items: %v\n", err)
		return summary, fmt.Errorf("error getting low stock items: %w", err)
	}

	// Get low stock count
	summary.LowStockCount = len(summary.LowStockItems)

	// Get top customers (limit to 5)
	summary.TopCustomers, err = r.GetTopCustomers(ctx, 5, days)
	if err != nil {
		fmt.Printf("Error getting top customers: %v\n", err)
		return summary, fmt.Errorf("error getting top customers: %w", err)
	}

	// Set period and last updated
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	summary.Period = fmt.Sprintf("Last %s - %s", startDate.Format("Jan 2"), endDate.Format("Jan 2"))
	summary.LastUpdated = time.Now()

	fmt.Println("Successfully retrieved dashboard summary")
	return summary, nil
}

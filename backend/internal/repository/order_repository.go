package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// OrderRepository handles database operations for orders and order items
type OrderRepository struct {
	db *sqlx.DB
}

// NewOrderRepository creates a new repository with the provided database connection
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// GetAll retrieves all orders from the database
func (r *OrderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	orders := []models.Order{}
	query := `SELECT * FROM orders ORDER BY order_date DESC`
	err := r.db.SelectContext(ctx, &orders, query)
	return orders, err
}

// GetByID retrieves an order by ID
func (r *OrderRepository) GetByID(ctx context.Context, id int) (models.Order, error) {
	var order models.Order
	query := `SELECT * FROM orders WHERE order_id = $1`
	err := r.db.GetContext(ctx, &order, query, id)
	if err == sql.ErrNoRows {
		return order, errors.New("order not found")
	}
	return order, err
}

// GetByCustomerID retrieves all orders for a specific customer
func (r *OrderRepository) GetByCustomerID(ctx context.Context, customerID int) ([]models.Order, error) {
	orders := []models.Order{}
	query := `SELECT * FROM orders WHERE customer_id = $1 ORDER BY order_date DESC`
	err := r.db.SelectContext(ctx, &orders, query, customerID)
	return orders, err
}

// Create inserts a new order into the database
func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	query := `
		INSERT INTO orders (
			customer_id, quotation_id, order_date, shipping_address, 
			status, total_amount, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING order_id, created_at, updated_at`

	err = tx.QueryRowContext(
		ctx,
		query,
		order.CustomerID,
		order.QuotationID,
		order.OrderDate,
		order.ShippingAddress,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&order.OrderID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		// Check for PostgreSQL-specific errors
		if pqErr, ok := err.(*pq.Error); ok {
			// 23505 is the PostgreSQL error code for unique_violation
			if pqErr.Code == "23505" {
				return ErrDuplicateKey
			}
		}
		return err
	}

	return tx.Commit()
}

// Update updates an existing order
func (r *OrderRepository) Update(ctx context.Context, order *models.Order) error {
	order.UpdatedAt = time.Now()

	query := `
		UPDATE orders SET
			customer_id = $1,
			quotation_id = $2,
			order_date = $3,
			shipping_address = $4,
			status = $5,
			total_amount = $6,
			updated_at = $7
		WHERE order_id = $8
		RETURNING updated_at`

	result := r.db.QueryRowContext(
		ctx,
		query,
		order.CustomerID,
		order.QuotationID,
		order.OrderDate,
		order.ShippingAddress,
		order.Status,
		order.TotalAmount,
		order.UpdatedAt,
		order.OrderID,
	)

	err := result.Scan(&order.UpdatedAt)
	if err == sql.ErrNoRows {
		return errors.New("order not found")
	}
	return err
}

// Delete removes an order by ID
func (r *OrderRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// First delete all order items associated with this order
	_, err = tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, id)
	if err != nil {
		return err
	}

	// Then delete the order itself
	result, err := tx.ExecContext(ctx, `DELETE FROM orders WHERE order_id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return tx.Commit()
}

// GetOrderItems retrieves all items for a specific order
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	items := []models.OrderItem{}
	query := `SELECT * FROM order_items WHERE order_id = $1`
	err := r.db.SelectContext(ctx, &items, query, orderID)
	return items, err
}

// CreateOrderItem inserts a new order item into the database
func (r *OrderRepository) CreateOrderItem(ctx context.Context, item *models.OrderItem) error {
	query := `
		INSERT INTO order_items (
			order_id, product_id, quantity, unit_price, discount
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING order_item_id, line_total`

	err := r.db.QueryRowContext(
		ctx,
		query,
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.UnitPrice,
		item.Discount,
	).Scan(&item.OrderItemID, &item.LineTotal)

	return err
}

// UpdateOrderItem updates an existing order item
func (r *OrderRepository) UpdateOrderItem(ctx context.Context, item *models.OrderItem) error {
	query := `
		UPDATE order_items SET
			order_id = $1,
			product_id = $2,
			quantity = $3,
			unit_price = $4,
			discount = $5
		WHERE order_item_id = $6
		RETURNING line_total`

	result := r.db.QueryRowContext(
		ctx,
		query,
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.UnitPrice,
		item.Discount,
		item.OrderItemID,
	)

	err := result.Scan(&item.LineTotal)
	if err == sql.ErrNoRows {
		return errors.New("order item not found")
	}
	return err
}

// DeleteOrderItem removes an order item by ID
func (r *OrderRepository) DeleteOrderItem(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM order_items WHERE order_item_id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order item not found")
	}

	return nil
}

// CreateOrderWithItems creates a new order with its items in a single transaction
func (r *OrderRepository) CreateOrderWithItems(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	// Insert the order first
	query := `
		INSERT INTO orders (
			customer_id, quotation_id, order_date, shipping_address, 
			status, total_amount, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING order_id, created_at, updated_at`

	err = tx.QueryRowContext(
		ctx,
		query,
		order.CustomerID,
		order.QuotationID,
		order.OrderDate,
		order.ShippingAddress,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&order.OrderID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return err
	}

	// Then insert all the items
	itemQuery := `
		INSERT INTO order_items (
			order_id, product_id, quantity, unit_price, discount
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING order_item_id, line_total`

	for i := range items {
		items[i].OrderID = order.OrderID
		err = tx.QueryRowContext(
			ctx,
			itemQuery,
			items[i].OrderID,
			items[i].ProductID,
			items[i].Quantity,
			items[i].UnitPrice,
			items[i].Discount,
		).Scan(&items[i].OrderItemID, &items[i].LineTotal)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateStatus updates only the status of an existing order
func (r *OrderRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"Pending":   true,
		"Shipped":   true,
		"Delivered": true,
		"Cancelled": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Get the current status of the order
	var currentStatus string
	err := r.db.QueryRowContext(ctx, "SELECT status FROM orders WHERE order_id = $1", id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("order not found")
		}
		return fmt.Errorf("failed to get current order status: %w", err)
	}

	// Validate status flow
	if currentStatus == "Cancelled" {
		return errors.New("cancelled orders cannot be updated")
	}

	if currentStatus == "Delivered" {
		return errors.New("delivered orders cannot be updated")
	}

	if currentStatus == "Shipped" && status == "Pending" {
		return errors.New("shipped orders cannot go back to pending status")
	}

	// Update the status in the database
	query := `
		UPDATE orders 
		SET status = $1, updated_at = NOW() 
		WHERE order_id = $2
		RETURNING *`

	var order models.Order
	err = r.db.QueryRowContext(ctx, query, status, id).Scan(
		&order.OrderID,
		&order.CustomerID,
		&order.QuotationID,
		&order.OrderDate,
		&order.ShippingAddress,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("order not found")
		}
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

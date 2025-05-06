package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// InventoryRepository handles database operations for inventory items
type InventoryRepository struct {
	db *sqlx.DB
}

// NewInventoryRepository creates a new repository with the provided database connection
func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{
		db: db,
	}
}

// GetAll retrieves all inventory items from the database
func (r *InventoryRepository) GetAll(ctx context.Context) ([]models.Inventory, error) {
	inventory := []models.Inventory{}
	query := `SELECT * FROM inventory ORDER BY inventory_id`
	err := r.db.SelectContext(ctx, &inventory, query)
	return inventory, err
}

// GetByID retrieves an inventory item by ID
func (r *InventoryRepository) GetByID(ctx context.Context, id int) (models.Inventory, error) {
	var inventory models.Inventory
	query := `SELECT * FROM inventory WHERE inventory_id = $1`
	err := r.db.GetContext(ctx, &inventory, query, id)
	if err == sql.ErrNoRows {
		return inventory, errors.New("inventory item not found")
	}
	return inventory, err
}

// GetByProductID retrieves inventory by product ID
func (r *InventoryRepository) GetByProductID(ctx context.Context, productID int) (models.Inventory, error) {
	var inventory models.Inventory
	query := `SELECT * FROM inventory WHERE product_id = $1`
	err := r.db.GetContext(ctx, &inventory, query, productID)
	if err == sql.ErrNoRows {
		return inventory, errors.New("inventory for product not found")
	}
	return inventory, err
}

// Create inserts a new inventory item into the database
func (r *InventoryRepository) Create(ctx context.Context, inventory *models.Inventory) error {
	query := `
		INSERT INTO inventory (
			product_id, current_stock, reorder_level, last_restock_date
		) VALUES (
			$1, $2, $3, $4
		) RETURNING inventory_id`

	err := r.db.QueryRowContext(
		ctx,
		query,
		inventory.ProductID,
		inventory.CurrentStock,
		inventory.ReorderLevel,
		inventory.LastRestockDate,
	).Scan(&inventory.InventoryID)

	if err != nil {
		// Check for PostgreSQL-specific errors
		if pqErr, ok := err.(*pq.Error); ok {
			// 23505 is the PostgreSQL error code for unique_violation
			if pqErr.Code == "23505" {
				return ErrDuplicateKey
			}
			// 23503 is the PostgreSQL error code for foreign_key_violation
			if pqErr.Code == "23503" {
				return errors.New("product not found")
			}
		}
	}

	return err
}

// Update updates an existing inventory item
func (r *InventoryRepository) Update(ctx context.Context, inventory *models.Inventory) error {
	query := `
		UPDATE inventory SET
			product_id = $1,
			current_stock = $2,
			reorder_level = $3,
			last_restock_date = $4
		WHERE inventory_id = $5`

	result, err := r.db.ExecContext(
		ctx,
		query,
		inventory.ProductID,
		inventory.CurrentStock,
		inventory.ReorderLevel,
		inventory.LastRestockDate,
		inventory.InventoryID,
	)

	if err != nil {
		// Check for unique constraint or foreign key violations
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return ErrDuplicateKey
			}
			if pqErr.Code == "23503" {
				return errors.New("product not found")
			}
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("inventory item not found")
	}

	return nil
}

// UpdateStock updates the current stock level and restock date
func (r *InventoryRepository) UpdateStock(ctx context.Context, inventoryID int, newStock int) error {
	now := time.Now()

	query := `
		UPDATE inventory SET
			current_stock = $1,
			last_restock_date = $2
		WHERE inventory_id = $3`

	result, err := r.db.ExecContext(ctx, query, newStock, now, inventoryID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("inventory item not found")
	}

	return nil
}

// Delete removes an inventory item by ID
func (r *InventoryRepository) Delete(ctx context.Context, id int) error {
	query := `
		WITH deleted AS (
			DELETE FROM inventory 
			WHERE inventory_id = $1 
			RETURNING inventory_id
		)
		SELECT COUNT(*) FROM deleted`

	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("inventory item not found")
	}

	return nil
}

// GetLowStockItems retrieves inventory items where current stock is at or below reorder level
func (r *InventoryRepository) GetLowStockItems(ctx context.Context) ([]models.Inventory, error) {
	inventory := []models.Inventory{}
	query := `
		SELECT * FROM inventory 
		WHERE current_stock <= reorder_level 
		ORDER BY (reorder_level - current_stock) DESC`
	
	err := r.db.SelectContext(ctx, &inventory, query)
	return inventory, err
}

// LowStockWithProductInfo combines product and inventory details for low stock items
type LowStockWithProductInfo struct {
	models.Inventory
	ProductName string  `db:"product_name" json:"product_name"`
	Price       float64 `db:"price" json:"price"`
}

// GetLowStockWithProductInfo retrieves low stock items with associated product info
func (r *InventoryRepository) GetLowStockWithProductInfo(ctx context.Context) ([]LowStockWithProductInfo, error) {
	items := []LowStockWithProductInfo{}
	query := `
		SELECT i.*, p.product_name, p.price 
		FROM inventory i
		JOIN products p ON i.product_id = p.product_id
		WHERE i.current_stock <= i.reorder_level
		ORDER BY (i.reorder_level - i.current_stock) DESC`
	
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
} 
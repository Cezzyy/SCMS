package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new repository with the provided database connection
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// GetAll retrieves all products from the database
func (r *ProductRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	products := []models.Product{}

	// We don't need the technical_specs::jsonb cast anymore since json.RawMessage handles it
	query := `
		SELECT * FROM products ORDER BY product_name
	`

	err := r.db.SelectContext(ctx, &products, query)
	if err != nil {
		return nil, errors.New("failed to retrieve products: " + err.Error())
	}

	return products, nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id int) (models.Product, error) {
	var product models.Product
	query := `SELECT * FROM products WHERE product_id = $1`

	err := r.db.GetContext(ctx, &product, query, id)
	if err == sql.ErrNoRows {
		return product, errors.New("product not found")
	}

	if err != nil {
		return product, errors.New("failed to retrieve product: " + err.Error())
	}

	return product, nil
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Ensure technical_specs is valid JSON for PostgreSQL
	if len(product.TechnicalSpecs) == 0 {
		// Initialize with empty JSON object
		product.TechnicalSpecs = json.RawMessage(`{}`)
	}

	// Use a placeholder for the JSONB column
	query := `
		INSERT INTO products (
			product_name, model, description, technical_specs, certifications,
			safety_standards, warranty_period, price, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4::jsonb, $5, $6, $7, $8, $9, $10
		) RETURNING product_id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		product.ProductName,
		product.Model,
		product.Description,
		product.TechnicalSpecs, // Already a json.RawMessage, no need to marshal
		product.Certifications,
		product.SafetyStandards,
		product.WarrantyPeriod,
		product.Price,
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(&product.ProductID, &product.CreatedAt, &product.UpdatedAt)

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

	return nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	// Ensure technical_specs is valid JSON for PostgreSQL
	if len(product.TechnicalSpecs) == 0 {
		// Initialize with empty JSON object
		product.TechnicalSpecs = json.RawMessage(`{}`)
	}

	query := `
		UPDATE products SET
			product_name = $1,
			model = $2,
			description = $3,
			technical_specs = $4::jsonb,
			certifications = $5,
			safety_standards = $6,
			warranty_period = $7,
			price = $8,
			updated_at = $9
		WHERE product_id = $10
		RETURNING updated_at`

	result := r.db.QueryRowContext(
		ctx,
		query,
		product.ProductName,
		product.Model,
		product.Description,
		product.TechnicalSpecs, // Already a json.RawMessage, no need to marshal
		product.Certifications,
		product.SafetyStandards,
		product.WarrantyPeriod,
		product.Price,
		product.UpdatedAt,
		product.ProductID,
	)

	err := result.Scan(&product.UpdatedAt)
	if err == sql.ErrNoRows {
		return errors.New("product not found")
	}

	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return ErrDuplicateKey
			}
		}
		return err
	}

	return nil
}

// Delete removes a product by ID
func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	// Using PostgreSQL's WITH clause for the deletion and getting count in one query
	query := `
		WITH deleted AS (
			DELETE FROM products 
			WHERE product_id = $1 
			RETURNING product_id
		)
		SELECT COUNT(*) FROM deleted`

	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("product not found")
	}

	return nil
}

// SearchProducts searches for products by name or description
func (r *ProductRepository) SearchProducts(ctx context.Context, term string) ([]models.Product, error) {
	products := []models.Product{}
	query := `
		SELECT * FROM products 
		WHERE product_name ILIKE $1 OR description ILIKE $1
		ORDER BY product_name`

	searchTerm := "%" + term + "%"
	err := r.db.SelectContext(ctx, &products, query, searchTerm)
	return products, err
}

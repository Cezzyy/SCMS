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

var (
	// ErrDuplicateKey is returned when a unique constraint is violated
	ErrDuplicateKey = errors.New("duplicate key value violates unique constraint")
)

// CustomerRepository handles database operations for customers
type CustomerRepository struct {
	db *sqlx.DB
}

// NewCustomerRepository creates a new repository with the provided database connection
func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{
		db: db,
	}
}

// GetAll retrieves all customers from the database
func (r *CustomerRepository) GetAll(ctx context.Context) ([]models.Customer, error) {
	customers := []models.Customer{}
	query := `SELECT * FROM customers ORDER BY company_name`
	err := r.db.SelectContext(ctx, &customers, query)
	return customers, err
}

// GetByID retrieves a customer by ID
func (r *CustomerRepository) GetByID(ctx context.Context, id int) (models.Customer, error) {
	var customer models.Customer
	query := `SELECT * FROM customers WHERE customer_id = $1`
	err := r.db.GetContext(ctx, &customer, query, id)
	if err == sql.ErrNoRows {
		return customer, errors.New("customer not found")
	}
	return customer, err
}

// Create inserts a new customer into the database
func (r *CustomerRepository) Create(ctx context.Context, customer *models.Customer) error {
	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now

	query := `
		INSERT INTO customers (
			company_name, industry, address, phone, email, website, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING customer_id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		customer.CompanyName,
		customer.Industry,
		customer.Address,
		customer.Phone,
		customer.Email,
		customer.Website,
		customer.CreatedAt,
		customer.UpdatedAt,
	).Scan(&customer.CustomerID, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		// Check for PostgreSQL-specific errors
		if pqErr, ok := err.(*pq.Error); ok {
			// 23505 is the PostgreSQL error code for unique_violation
			if pqErr.Code == "23505" {
				return ErrDuplicateKey
			}
		}
	}

	return err
}

// Update updates an existing customer
func (r *CustomerRepository) Update(ctx context.Context, customer *models.Customer) error {
	customer.UpdatedAt = time.Now()

	query := `
		UPDATE customers SET
			company_name = $1,
			industry = $2,
			address = $3,
			phone = $4,
			email = $5,
			website = $6,
			updated_at = $7
		WHERE customer_id = $8
		RETURNING updated_at`

	result := r.db.QueryRowContext(
		ctx,
		query,
		customer.CompanyName,
		customer.Industry,
		customer.Address,
		customer.Phone,
		customer.Email,
		customer.Website,
		customer.UpdatedAt,
		customer.CustomerID,
	)

	err := result.Scan(&customer.UpdatedAt)
	if err == sql.ErrNoRows {
		return errors.New("customer not found")
	}
	return err
}

// Delete removes a customer by ID
func (r *CustomerRepository) Delete(ctx context.Context, id int) error {
	// Using PostgreSQL's WITH clause for the deletion and getting count in one query
	query := `
		WITH deleted AS (
			DELETE FROM customers 
			WHERE customer_id = $1 
			RETURNING customer_id
		)
		SELECT COUNT(*) FROM deleted`

	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("customer not found")
	}

	return nil
}

// SearchCustomers searches for customers by company name using PostgreSQL's ILIKE
func (r *CustomerRepository) SearchCustomers(ctx context.Context, term string) ([]models.Customer, error) {
	customers := []models.Customer{}
	query := `SELECT * FROM customers WHERE company_name ILIKE $1 ORDER BY company_name`
	err := r.db.SelectContext(ctx, &customers, query, "%"+term+"%")
	return customers, err
}

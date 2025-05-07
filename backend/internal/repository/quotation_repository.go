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

// QuotationRepository handles database operations for quotations and quotation items
type QuotationRepository struct {
	db *sqlx.DB
}

// NewQuotationRepository creates a new repository with the provided database connection
func NewQuotationRepository(db *sqlx.DB) *QuotationRepository {
	return &QuotationRepository{
		db: db,
	}
}

// GetAll retrieves all quotations from the database
func (r *QuotationRepository) GetAll(ctx context.Context) ([]models.Quotation, error) {
	quotations := []models.Quotation{}
	query := `SELECT * FROM quotations ORDER BY quote_date DESC`
	err := r.db.SelectContext(ctx, &quotations, query)
	return quotations, err
}

// GetByID retrieves a quotation by ID
func (r *QuotationRepository) GetByID(ctx context.Context, id int) (models.Quotation, error) {
	var quotation models.Quotation
	query := `SELECT * FROM quotations WHERE quotation_id = $1`
	err := r.db.GetContext(ctx, &quotation, query, id)
	if err == sql.ErrNoRows {
		return quotation, errors.New("quotation not found")
	}
	return quotation, err
}

// GetByCustomerID retrieves all quotations for a specific customer
func (r *QuotationRepository) GetByCustomerID(ctx context.Context, customerID int) ([]models.Quotation, error) {
	quotations := []models.Quotation{}
	query := `SELECT * FROM quotations WHERE customer_id = $1 ORDER BY quote_date DESC`
	err := r.db.SelectContext(ctx, &quotations, query, customerID)
	return quotations, err
}

// Create inserts a new quotation into the database
func (r *QuotationRepository) Create(ctx context.Context, quotation *models.Quotation) error {
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
	quotation.CreatedAt = now
	quotation.UpdatedAt = now

	query := `
		INSERT INTO quotations (
			customer_id, quote_date, validity_date, status, 
			total_amount, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING quotation_id, created_at, updated_at`

	err = tx.QueryRowContext(
		ctx,
		query,
		quotation.CustomerID,
		quotation.QuoteDate,
		quotation.ValidityDate,
		quotation.Status,
		quotation.TotalAmount,
		quotation.CreatedAt,
		quotation.UpdatedAt,
	).Scan(&quotation.QuotationID, &quotation.CreatedAt, &quotation.UpdatedAt)

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

// Update updates an existing quotation
func (r *QuotationRepository) Update(ctx context.Context, quotation *models.Quotation) error {
	quotation.UpdatedAt = time.Now()

	query := `
		UPDATE quotations SET
			customer_id = $1,
			quote_date = $2,
			validity_date = $3,
			status = $4,
			total_amount = $5,
			updated_at = $6
		WHERE quotation_id = $7
		RETURNING updated_at`

	result := r.db.QueryRowContext(
		ctx,
		query,
		quotation.CustomerID,
		quotation.QuoteDate,
		quotation.ValidityDate,
		quotation.Status,
		quotation.TotalAmount,
		quotation.UpdatedAt,
		quotation.QuotationID,
	)

	err := result.Scan(&quotation.UpdatedAt)
	if err == sql.ErrNoRows {
		return errors.New("quotation not found")
	}
	return err
}

// Delete removes a quotation by ID
func (r *QuotationRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// First delete all quotation items associated with this quotation
	_, err = tx.ExecContext(ctx, `DELETE FROM quotation_items WHERE quotation_id = $1`, id)
	if err != nil {
		return err
	}

	// Then delete the quotation itself
	result, err := tx.ExecContext(ctx, `DELETE FROM quotations WHERE quotation_id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quotation not found")
	}

	return tx.Commit()
}

// GetQuotationItems retrieves all items for a specific quotation
func (r *QuotationRepository) GetQuotationItems(ctx context.Context, quotationID int) ([]models.QuotationItem, error) {
	items := []models.QuotationItem{}
	query := `SELECT * FROM quotation_items WHERE quotation_id = $1`
	err := r.db.SelectContext(ctx, &items, query, quotationID)
	return items, err
}

// CreateQuotationItem inserts a new quotation item into the database
func (r *QuotationRepository) CreateQuotationItem(ctx context.Context, item *models.QuotationItem) error {
	query := `
		INSERT INTO quotation_items (
			quotation_id, product_id, quantity, unit_price, discount, line_total
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING quotation_item_id`

	err := r.db.QueryRowContext(
		ctx,
		query,
		item.QuotationID,
		item.ProductID,
		item.Quantity,
		item.UnitPrice,
		item.Discount,
		item.LineTotal,
	).Scan(&item.QuotationItemID)

	return err
}

// UpdateQuotationItem updates an existing quotation item
func (r *QuotationRepository) UpdateQuotationItem(ctx context.Context, item *models.QuotationItem) error {
	query := `
		UPDATE quotation_items SET
			quotation_id = $1,
			product_id = $2,
			quantity = $3,
			unit_price = $4,
			discount = $5,
			line_total = $6
		WHERE quotation_item_id = $7`

	result, err := r.db.ExecContext(
		ctx,
		query,
		item.QuotationID,
		item.ProductID,
		item.Quantity,
		item.UnitPrice,
		item.Discount,
		item.LineTotal,
		item.QuotationItemID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quotation item not found")
	}

	return nil
}

// DeleteQuotationItem removes a quotation item by ID
func (r *QuotationRepository) DeleteQuotationItem(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM quotation_items WHERE quotation_item_id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quotation item not found")
	}

	return nil
}

// GetFullQuotation retrieves a quotation along with all its items
func (r *QuotationRepository) GetFullQuotation(ctx context.Context, id int) (models.Quotation, []models.QuotationItem, error) {
	// Get the quotation
	quotation, err := r.GetByID(ctx, id)
	if err != nil {
		return quotation, nil, err
	}

	// Get the quotation items
	items, err := r.GetQuotationItems(ctx, id)
	if err != nil {
		return quotation, nil, err
	}

	return quotation, items, nil
}

// CreateQuotationWithItems creates a new quotation with its items in a single transaction
func (r *QuotationRepository) CreateQuotationWithItems(ctx context.Context, quotation *models.Quotation, items []models.QuotationItem) error {
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
	quotation.CreatedAt = now
	quotation.UpdatedAt = now

	// Insert the quotation first
	query := `
		INSERT INTO quotations (
			customer_id, quote_date, validity_date, status, 
			total_amount, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING quotation_id, created_at, updated_at`

	err = tx.QueryRowContext(
		ctx,
		query,
		quotation.CustomerID,
		quotation.QuoteDate,
		quotation.ValidityDate,
		quotation.Status,
		quotation.TotalAmount,
		quotation.CreatedAt,
		quotation.UpdatedAt,
	).Scan(&quotation.QuotationID, &quotation.CreatedAt, &quotation.UpdatedAt)

	if err != nil {
		return err
	}

	// Then insert all the items
	itemQuery := `
		INSERT INTO quotation_items (
			quotation_id, product_id, quantity, unit_price, discount, line_total
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING quotation_item_id`

	for i := range items {
		items[i].QuotationID = quotation.QuotationID
		err = tx.QueryRowContext(
			ctx,
			itemQuery,
			items[i].QuotationID,
			items[i].ProductID,
			items[i].Quantity,
			items[i].UnitPrice,
			items[i].Discount,
			items[i].LineTotal,
		).Scan(&items[i].QuotationItemID)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

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

// ContactRepository handles database operations for contacts
type ContactRepository struct {
	db *sqlx.DB
}

// NewContactRepository creates a new repository with the provided database connection
func NewContactRepository(db *sqlx.DB) *ContactRepository {
	return &ContactRepository{
		db: db,
	}
}

// GetAll retrieves all contacts from the database
func (r *ContactRepository) GetAll(ctx context.Context) ([]models.Contact, error) {
	contacts := []models.Contact{}
	query := `SELECT * FROM contacts ORDER BY last_name, first_name`
	err := r.db.SelectContext(ctx, &contacts, query)
	return contacts, err
}

// GetByID retrieves a contact by ID
func (r *ContactRepository) GetByID(ctx context.Context, id int) (models.Contact, error) {
	var contact models.Contact
	query := `SELECT * FROM contacts WHERE contact_id = $1`
	err := r.db.GetContext(ctx, &contact, query, id)
	if err == sql.ErrNoRows {
		return contact, errors.New("contact not found")
	}
	return contact, err
}

// GetByCustomerID retrieves all contacts for a specific customer
func (r *ContactRepository) GetByCustomerID(ctx context.Context, customerID int) ([]models.Contact, error) {
	contacts := []models.Contact{}
	query := `SELECT * FROM contacts WHERE customer_id = $1 ORDER BY last_name, first_name`
	err := r.db.SelectContext(ctx, &contacts, query, customerID)
	return contacts, err
}

// Create inserts a new contact into the database
func (r *ContactRepository) Create(ctx context.Context, contact *models.Contact) error {
	now := time.Now()
	contact.CreatedAt = now
	contact.UpdatedAt = now

	query := `
		INSERT INTO contacts (
			customer_id, first_name, last_name, position, phone, email, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING contact_id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		contact.CustomerID,
		contact.FirstName,
		contact.LastName,
		contact.Position,
		contact.Phone,
		contact.Email,
		contact.CreatedAt,
		contact.UpdatedAt,
	).Scan(&contact.ContactID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		// Check for PostgreSQL-specific errors
		if pqErr, ok := err.(*pq.Error); ok {
			// 23505 is the PostgreSQL error code for unique_violation
			if pqErr.Code == "23505" {
				return ErrDuplicateKey
			}
			// 23503 is the PostgreSQL error code for foreign_key_violation
			if pqErr.Code == "23503" {
				return errors.New("customer not found")
			}
		}
	}

	return err
}

// Update updates an existing contact
func (r *ContactRepository) Update(ctx context.Context, contact *models.Contact) error {
	contact.UpdatedAt = time.Now()

	query := `
		UPDATE contacts SET
			customer_id = $1,
			first_name = $2,
			last_name = $3,
			position = $4,
			phone = $5,
			email = $6,
			updated_at = $7
		WHERE contact_id = $8
		RETURNING updated_at`

	result := r.db.QueryRowContext(
		ctx,
		query,
		contact.CustomerID,
		contact.FirstName,
		contact.LastName,
		contact.Position,
		contact.Phone,
		contact.Email,
		contact.UpdatedAt,
		contact.ContactID,
	)

	err := result.Scan(&contact.UpdatedAt)
	if err == sql.ErrNoRows {
		return errors.New("contact not found")
	}

	if err != nil {
		// Check for foreign key violation
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" {
				return errors.New("customer not found")
			}
		}
	}

	return err
}

// Delete removes a contact by ID
func (r *ContactRepository) Delete(ctx context.Context, id int) error {
	// Using PostgreSQL's WITH clause for the deletion and getting count in one query
	query := `
		WITH deleted AS (
			DELETE FROM contacts 
			WHERE contact_id = $1 
			RETURNING contact_id
		)
		SELECT COUNT(*) FROM deleted`

	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("contact not found")
	}

	return nil
}

// SearchContacts searches for contacts by name using PostgreSQL's ILIKE
func (r *ContactRepository) SearchContacts(ctx context.Context, term string) ([]models.Contact, error) {
	contacts := []models.Contact{}
	// Using PostgreSQL's CONCAT and ILIKE for case-insensitive search
	query := `
		SELECT * FROM contacts 
		WHERE CONCAT(first_name, ' ', last_name) ILIKE $1 
		ORDER BY last_name, first_name`
	err := r.db.SelectContext(ctx, &contacts, query, "%"+term+"%")
	return contacts, err
}

// CheckEmailExists checks if an email already exists
func (r *ContactRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM contacts WHERE email = $1)`
	err := r.db.GetContext(ctx, &exists, query, email)
	return exists, err
}

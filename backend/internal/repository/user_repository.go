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

// UserRepository handles database operations for users
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new repository with the provided database connection
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetAll retrieves all users from the database
func (r *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	users := []models.User{}
	query := `SELECT * FROM users ORDER BY email`
	err := r.db.SelectContext(ctx, &users, query)
	return users, err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE user_id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return user, errors.New("user not found")
	}
	return user, err
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return user, errors.New("user not found")
	}
	return user, err
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (
			password_hash, role, first_name, last_name, 
			email, phone, department, position, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING user_id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.PasswordHash,
		user.Role,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Phone,
		user.Department,
		user.Position,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)

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

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			role = $1,
			first_name = $2,
			last_name = $3,
			email = $4,
			phone = $5,
			department = $6,
			position = $7,
			updated_at = $8
		WHERE user_id = $9
		RETURNING updated_at`

	result := r.db.QueryRowContext(
		ctx,
		query,
		user.Role,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Phone,
		user.Department,
		user.Position,
		user.UpdatedAt,
		user.UserID,
	)

	err := result.Scan(&user.UpdatedAt)
	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}

	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return ErrDuplicateKey
			}
		}
	}

	return err
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int, passwordHash string) error {
	now := time.Now()

	query := `
		UPDATE users SET
			password_hash = $1,
			updated_at = $2
		WHERE user_id = $3
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, passwordHash, now, userID).Scan(&updatedAt)

	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}

	return err
}

// UpdateLastLogin updates a user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	now := time.Now()

	query := `
		UPDATE users SET
			last_login = $1
		WHERE user_id = $2`

	_, err := r.db.ExecContext(ctx, query, now, userID)
	return err
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	// Using PostgreSQL's WITH clause for the deletion and getting count in one query
	query := `
		WITH deleted AS (
			DELETE FROM users 
			WHERE user_id = $1 
			RETURNING user_id
		)
		SELECT COUNT(*) FROM deleted`

	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("user not found")
	}

	return nil
}

// SearchUsers searches for users by name or email
func (r *UserRepository) SearchUsers(ctx context.Context, term string) ([]models.User, error) {
	users := []models.User{}

	// Using PostgreSQL's ILIKE for case-insensitive search on multiple fields
	query := `
		SELECT * FROM users 
		WHERE CONCAT(first_name, ' ', last_name) ILIKE $1
		   OR email ILIKE $1
		ORDER BY email`

	err := r.db.SelectContext(ctx, &users, query, "%"+term+"%")
	return users, err
}

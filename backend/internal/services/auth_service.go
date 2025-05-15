package services

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// LoginRequest contains the credentials submitted by the user
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse contains user data and session information
type AuthResponse struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login authenticates a user and returns a session
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login time
	s.userRepo.UpdateLastLogin(ctx, user.UserID)

	// Create simple session ID (in a real app, this would be more secure)
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	return &AuthResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}, nil
}

// Helper function to generate a simple session ID
func generateSessionID() string {
	// In a real app, use a more secure method like crypto/rand
	return "sess_" + time.Now().Format("20060102150405") + "_" + strconv.Itoa(time.Now().Nanosecond())
}

// HashPassword hashes a password for storage
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

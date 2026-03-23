package user

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service defines business logic operations for users
type Service interface {
	CreateUser(req CreateUserRequest) (*User, error)
	GetUserByID(id uuid.UUID) (*User, error)
	GetAllUsers() ([]*User, error)
	GetUserByEmail(email string) (*User, error)
	MarkEmailVerified(id uuid.UUID) error
	UpdatePassword(id uuid.UUID, plainPassword string) error
}

type service struct {
	repo Repository
}

// NewService creates a new user service with the given repository
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateUser validates and persists a new user
func (s *service) CreateUser(req CreateUserRequest) (*User, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, fmt.Errorf("password is required")
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &User{
		ID:                 uuid.New(),
		Name:               strings.TrimSpace(req.Name),
		Email:              strings.ToLower(strings.TrimSpace(req.Email)),
		EmailVerified:      false,
		EmailVerifiedAt:    nil,
		PasswordHash:       string(passwordHash),
		ForecastEfficiency: 0.8,
		CreatedAt:          time.Now().UTC(),
	}

	if err := s.repo.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByID retrieves a user by their UUID
func (s *service) GetUserByID(id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(id)
}

// GetAllUsers returns all registered users
func (s *service) GetAllUsers() ([]*User, error) {
	return s.repo.GetAllUsers()
}

// GetUserByEmail retrieves a user by email.
func (s *service) GetUserByEmail(email string) (*User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email is required")
	}

	return s.repo.GetUserByEmail(strings.ToLower(strings.TrimSpace(email)))
}

// MarkEmailVerified marks one user email as verified.
func (s *service) MarkEmailVerified(id uuid.UUID) error {
	return s.repo.MarkEmailVerified(id)
}

// UpdatePassword hashes and updates the password for one user.
func (s *service) UpdatePassword(id uuid.UUID, plainPassword string) error {
	if len(strings.TrimSpace(plainPassword)) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.repo.UpdatePassword(id, string(passwordHash))
}

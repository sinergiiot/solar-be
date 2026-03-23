package user

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Repository defines data access methods for users
type Repository interface {
	CreateUser(u *User) error
	GetUserByID(id uuid.UUID) (*User, error)
	GetAllUsers() ([]*User, error)
	GetUserByEmail(email string) (*User, error)
	MarkEmailVerified(id uuid.UUID) error
	UpdatePassword(id uuid.UUID, passwordHash string) error
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository backed by the given database
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// CreateUser inserts a new user into the database
func (r *repository) CreateUser(u *User) error {
	query := `
		INSERT INTO users (id, name, email, email_verified, email_verified_at, password_hash, forecast_efficiency, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, u.ID, u.Name, u.Email, u.EmailVerified, u.EmailVerifiedAt, u.PasswordHash, u.ForecastEfficiency, u.CreatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetUserByID fetches a single user by their UUID
func (r *repository) GetUserByID(id uuid.UUID) (*User, error) {
	query := `SELECT id, name, email, email_verified, email_verified_at, password_hash, forecast_efficiency, created_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)

	u := &User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.EmailVerifiedAt, &u.PasswordHash, &u.ForecastEfficiency, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetAllUsers returns every user stored in the database
func (r *repository) GetAllUsers() ([]*User, error) {
	query := `SELECT id, name, email, email_verified, email_verified_at, password_hash, forecast_efficiency, created_at FROM users ORDER BY created_at ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.EmailVerifiedAt, &u.PasswordHash, &u.ForecastEfficiency, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserByEmail fetches a single user by email.
func (r *repository) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, name, email, email_verified, email_verified_at, password_hash, forecast_efficiency, created_at FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	u := &User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.EmailVerifiedAt, &u.PasswordHash, &u.ForecastEfficiency, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return u, nil
}

// MarkEmailVerified sets email verification flags for one user.
func (r *repository) MarkEmailVerified(id uuid.UUID) error {
	query := `UPDATE users SET email_verified = TRUE, email_verified_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}

	return nil
}

// UpdatePassword updates the password hash for one user.
func (r *repository) UpdatePassword(id uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := r.db.Exec(query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	return nil
}

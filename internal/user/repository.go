package user

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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
	UpdateBranding(id uuid.UUID, companyName, logoURL string) error
	// E5-T6: ESG Share
	SetESGShareToken(id uuid.UUID, token string, enabled bool) error
	GetUserByESGShareToken(token string) (*User, error)
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository backed by the given database
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// scanUser is a helper to scan the standard user row columns.
func scanUser(s interface {
	Scan(...any) error
}) (*User, error) {
	u := &User{}
	var shareToken sql.NullString
	err := s.Scan(
		&u.ID, &u.Name, &u.Email, &u.Role,
		&u.EmailVerified, &u.EmailVerifiedAt,
		&u.PasswordHash, &u.ForecastEfficiency,
		&u.CompanyLogoURL, &u.CompanyName,
		&shareToken, &u.ESGShareEnabled,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.ESGShareToken = shareToken.String
	return u, nil
}

const selectUserCols = `id, name, email, role, email_verified, email_verified_at,
  password_hash, forecast_efficiency, company_logo_url, company_name,
  esg_share_token, esg_share_enabled, created_at, updated_at`

// CreateUser inserts a new user into the database
func (r *repository) CreateUser(u *User) error {
	query := `
		INSERT INTO users (id, name, email, role, email_verified, email_verified_at, password_hash, forecast_efficiency, company_logo_url, company_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(query, u.ID, u.Name, u.Email, u.Role, u.EmailVerified, u.EmailVerifiedAt, u.PasswordHash, u.ForecastEfficiency, u.CompanyLogoURL, u.CompanyName, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetUserByID fetches a single user by their UUID
func (r *repository) GetUserByID(id uuid.UUID) (*User, error) {
	query := `SELECT ` + selectUserCols + ` FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetAllUsers returns every user stored in the database
func (r *repository) GetAllUsers() ([]*User, error) {
	query := `SELECT ` + selectUserCols + ` FROM users ORDER BY created_at ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserByEmail fetches a single user by email.
func (r *repository) GetUserByEmail(email string) (*User, error) {
	query := `SELECT ` + selectUserCols + ` FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)
	u, err := scanUser(row)
	if err != nil {
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

// UpdateBranding updates the company logo and name for a user.
func (r *repository) UpdateBranding(id uuid.UUID, companyName, logoURL string) error {
	query := `UPDATE users SET company_name = $1, company_logo_url = $2 WHERE id = $3`
	_, err := r.db.Exec(query, companyName, logoURL, id)
	if err != nil {
		return fmt.Errorf("update branding: %w", err)
	}
	return nil
}

// SetESGShareToken sets or clears the ESG share token for a user.
func (r *repository) SetESGShareToken(id uuid.UUID, token string, enabled bool) error {
	var query string
	var args []any
	if token == "" {
		query = `UPDATE users SET esg_share_token = NULL, esg_share_enabled = $1 WHERE id = $2`
		args = []any{enabled, id}
	} else {
		query = `UPDATE users SET esg_share_token = $1, esg_share_enabled = $2 WHERE id = $3`
		args = []any{token, enabled, id}
	}
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("set esg share token: %w", err)
	}
	return nil
}

// GetUserByESGShareToken retrieves a user by their public ESG share token.
func (r *repository) GetUserByESGShareToken(token string) (*User, error) {
	query := `SELECT ` + selectUserCols + ` FROM users WHERE esg_share_token = $1 AND esg_share_enabled = TRUE`
	row := r.db.QueryRow(query, token)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("get user by esg share token: %w", err)
	}
	return u, nil
}

// GenerateShareToken returns a secure random 32-byte hex token.
func GenerateShareToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

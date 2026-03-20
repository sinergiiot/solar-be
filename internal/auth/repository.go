package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// refreshTokenRepository handles DB operations for refresh tokens.
type refreshTokenRepository struct {
	db *sql.DB
}

func newRefreshTokenRepository(db *sql.DB) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// refreshTokenRow represents one row from the refresh_tokens table.
type refreshTokenRow struct {
	userID    uuid.UUID
	expiresAt time.Time
	revoked   bool
}

// generateToken creates a cryptographically secure random 32-byte hex token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Create stores a new refresh token for the given user and returns its value.
func (r *refreshTokenRepository) Create(userID uuid.UUID, expiresAt time.Time) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	_, err = r.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Find retrieves the metadata for a given refresh token string.
func (r *refreshTokenRepository) Find(token string) (*refreshTokenRow, error) {
	var row refreshTokenRow
	err := r.db.QueryRow(
		`SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token = $1`,
		token,
	).Scan(&row.userID, &row.expiresAt, &row.revoked)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Revoke marks a single token as revoked.
func (r *refreshTokenRepository) Revoke(token string) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`, token)
	return err
}

// RevokeAllForUser revokes every refresh token belonging to one user.
func (r *refreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`, userID)
	return err
}

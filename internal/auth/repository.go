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

// emailVerificationRepository handles DB operations for email verification OTP.
type emailVerificationRepository struct {
	db *sql.DB
}

func newRefreshTokenRepository(db *sql.DB) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func newEmailVerificationRepository(db *sql.DB) *emailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

type passwordResetRepository struct {
	db *sql.DB
}

func newPasswordResetRepository(db *sql.DB) *passwordResetRepository {
	return &passwordResetRepository{db: db}
}

// refreshTokenRow represents one row from the refresh_tokens table.
type refreshTokenRow struct {
	userID    uuid.UUID
	expiresAt time.Time
	revoked   bool
}

type emailVerificationRow struct {
	userID    uuid.UUID
	expiresAt time.Time
	usedAt    sql.NullTime
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

// Create stores a short-lived verification code for one user.
func (r *emailVerificationRepository) Create(userID uuid.UUID, code string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO email_verification_codes (user_id, code, expires_at) VALUES ($1, $2, $3)`,
		userID, code, expiresAt,
	)
	return err
}

// FindLatestByUserAndCode reads the latest code record for one user and code.
func (r *emailVerificationRepository) FindLatestByUserAndCode(userID uuid.UUID, code string) (*emailVerificationRow, error) {
	row := &emailVerificationRow{}
	err := r.db.QueryRow(
		`SELECT user_id, expires_at, used_at
		 FROM email_verification_codes
		 WHERE user_id = $1 AND code = $2
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID,
		code,
	).Scan(&row.userID, &row.expiresAt, &row.usedAt)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// InvalidateAllByUser marks all active codes for one user as used.
func (r *emailVerificationRepository) InvalidateAllByUser(userID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE email_verification_codes SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`, userID)
	return err
}

// MarkCodeUsed marks matching active code as consumed.
func (r *emailVerificationRepository) MarkCodeUsed(userID uuid.UUID, code string) error {
	_, err := r.db.Exec(
		`UPDATE email_verification_codes SET used_at = NOW()
		 WHERE user_id = $1 AND code = $2 AND used_at IS NULL`,
		userID,
		code,
	)
	return err
}

// Create stores a short-lived reset code for one user.
func (r *passwordResetRepository) Create(userID uuid.UUID, code string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO password_reset_codes (user_id, code, expires_at) VALUES ($1, $2, $3)`,
		userID, code, expiresAt,
	)
	return err
}

// FindLatestByUserAndCode reads the latest code record for one user and code.
func (r *passwordResetRepository) FindLatestByUserAndCode(userID uuid.UUID, code string) (*emailVerificationRow, error) {
	row := &emailVerificationRow{}
	err := r.db.QueryRow(
		`SELECT user_id, expires_at, used_at
		 FROM password_reset_codes
		 WHERE user_id = $1 AND code = $2
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID,
		code,
	).Scan(&row.userID, &row.expiresAt, &row.usedAt)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// InvalidateAllByUser marks all active codes for one user as used.
func (r *passwordResetRepository) InvalidateAllByUser(userID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE password_reset_codes SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`, userID)
	return err
}

// MarkCodeUsed marks matching active code as consumed.
func (r *passwordResetRepository) MarkCodeUsed(userID uuid.UUID, code string) error {
	_, err := r.db.Exec(
		`UPDATE password_reset_codes SET used_at = NOW()
		 WHERE user_id = $1 AND code = $2 AND used_at IS NULL`,
		userID,
		code,
	)
	return err
}

package apikey

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Create(key *UserAPIKey) error
	GetByID(id uuid.UUID) (*UserAPIKey, error)
	GetByHash(hash string) (*UserAPIKey, error)
	ListByUserID(userID uuid.UUID) ([]*UserAPIKey, error)
	Delete(id uuid.UUID, userID uuid.UUID) error
	UpdateLastUsed(id uuid.UUID) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(key *UserAPIKey) error {
	query := `
		INSERT INTO user_api_keys (id, user_id, name, api_key_hash, api_key_preview, is_active, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, key.ID, key.UserID, key.Name, key.APIKeyHash, key.APIKeyPreview, key.IsActive, key.ExpiresAt, key.CreatedAt)
	return err
}

func (r *repository) GetByID(id uuid.UUID) (*UserAPIKey, error) {
	query := `SELECT id, user_id, name, api_key_hash, api_key_preview, is_active, last_used_at, expires_at, created_at FROM user_api_keys WHERE id = $1`
	key := &UserAPIKey{}
	err := r.db.QueryRow(query, id).Scan(&key.ID, &key.UserID, &key.Name, &key.APIKeyHash, &key.APIKeyPreview, &key.IsActive, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("api key not found")
	}
	return key, err
}

func (r *repository) GetByHash(hash string) (*UserAPIKey, error) {
	query := `SELECT id, user_id, name, api_key_hash, api_key_preview, is_active, last_used_at, expires_at, created_at FROM user_api_keys WHERE api_key_hash = $1 AND is_active = TRUE`
	key := &UserAPIKey{}
	err := r.db.QueryRow(query, hash).Scan(&key.ID, &key.UserID, &key.Name, &key.APIKeyHash, &key.APIKeyPreview, &key.IsActive, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid or inactive api key")
	}
	return key, err
}

func (r *repository) ListByUserID(userID uuid.UUID) ([]*UserAPIKey, error) {
	query := `SELECT id, user_id, name, api_key_preview, is_active, last_used_at, expires_at, created_at FROM user_api_keys WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*UserAPIKey
	for rows.Next() {
		k := &UserAPIKey{}
		err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.APIKeyPreview, &k.IsActive, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *repository) Delete(id uuid.UUID, userID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM user_api_keys WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *repository) UpdateLastUsed(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE user_api_keys SET last_used_at = NOW() WHERE id = $1`, id)
	return err
}

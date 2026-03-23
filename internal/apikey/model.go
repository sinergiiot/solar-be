package apikey

import (
	"time"

	"github.com/google/uuid"
)

type UserAPIKey struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Name          string     `json:"name"`
	APIKeyHash    string     `json:"-"`
	APIKeyPreview string     `json:"api_key_preview"`
	IsActive      bool       `json:"is_active"`
	LastUsedAt    *time.Time `json:"last_used_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type CreateAPIKeyResponse struct {
	ID     uuid.UUID `json:"id"`
	APIKey string    `json:"api_key"` // Only shown once
}

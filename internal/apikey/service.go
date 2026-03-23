package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
)

type Service interface {
	CreateKey(userID uuid.UUID, name string, expiresAt *time.Time) (*CreateAPIKeyResponse, error)
	ListKeys(userID uuid.UUID) ([]*UserAPIKey, error)
	DeleteKey(id uuid.UUID, userID uuid.UUID) error
	ValidateKey(keyString string) (uuid.UUID, string, error)
}

type service struct {
	repo    Repository
	userSvc user.Service
}

func NewService(repo Repository, userSvc user.Service) Service {
	return &service{repo: repo, userSvc: userSvc}
}

func (s *service) CreateKey(userID uuid.UUID, name string, expiresAt *time.Time) (*CreateAPIKeyResponse, error) {
	rawKey, err := generateRandomKey(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate raw key: %w", err)
	}

	keyString := "sk_live_" + rawKey
	hash := hashKey(keyString)
	preview := keyString[:12] + "..." + keyString[len(keyString)-4:]

	apiKey := &UserAPIKey{
		ID:            uuid.New(),
		UserID:        userID,
		Name:          name,
		APIKeyHash:    hash,
		APIKeyPreview: preview,
		IsActive:      true,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.Create(apiKey); err != nil {
		return nil, fmt.Errorf("failed to save api key: %w", err)
	}

	return &CreateAPIKeyResponse{
		ID:     apiKey.ID,
		APIKey: keyString,
	}, nil
}

func (s *service) ListKeys(userID uuid.UUID) ([]*UserAPIKey, error) {
	return s.repo.ListByUserID(userID)
}

func (s *service) DeleteKey(id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

func (s *service) ValidateKey(keyString string) (uuid.UUID, string, error) {
	hash := hashKey(keyString)
	apiKey, err := s.repo.GetByHash(hash)
	if err != nil {
		return uuid.Nil, "", err
	}

	if apiKey.ExpiresAt != nil && time.Now().UTC().After(*apiKey.ExpiresAt) {
		return uuid.Nil, "", fmt.Errorf("api key expired")
	}

	// Fetch user role
	u, err := s.userSvc.GetUserByID(apiKey.UserID)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("user not found")
	}

	if err := s.repo.UpdateLastUsed(apiKey.ID); err != nil {
		// Log error but don't fail validation
		fmt.Printf("failed to update api key last used: %v\n", err)
	}

	return apiKey.UserID, u.Role, nil
}

func generateRandomKey(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashKey(key string) string {
	h := sha256.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

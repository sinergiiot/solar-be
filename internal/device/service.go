package device

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service defines business operations for device management and ingestion.
type Service interface {
	CreateDevice(userID uuid.UUID, req CreateDeviceRequest) (*CreateDeviceResponse, error)
	ListDevices(userID uuid.UUID) ([]*Device, error)
	GetHeartbeatSummary(userID uuid.UUID) (*HeartbeatSummary, error)
	UpdateDevice(userID uuid.UUID, deviceID uuid.UUID, req UpdateDeviceRequest) (*Device, error)
	DeleteDevice(userID uuid.UUID, deviceID uuid.UUID) error
	RotateDeviceKey(userID uuid.UUID, deviceID uuid.UUID) (*RotateKeyResponse, error)
	IngestTelemetry(apiKey string, req IngestTelemetryRequest) (*IngestTelemetryResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a new device service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateDevice validates input and registers one device with a generated api key.
func (s *service) CreateDevice(userID uuid.UUID, req CreateDeviceRequest) (*CreateDeviceResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	externalID := strings.TrimSpace(req.ExternalID)
	if externalID == "" {
		return nil, fmt.Errorf("external_id is required")
	}

	var solarProfileID *uuid.UUID
	if req.SolarProfileID != nil && strings.TrimSpace(*req.SolarProfileID) != "" {
		parsedProfileID, err := uuid.Parse(strings.TrimSpace(*req.SolarProfileID))
		if err != nil {
			return nil, fmt.Errorf("solar_profile_id is invalid")
		}

		owned, err := s.repo.IsSolarProfileOwnedByUser(userID, parsedProfileID)
		if err != nil {
			return nil, err
		}
		if !owned {
			return nil, fmt.Errorf("solar_profile_id does not belong to this user")
		}

		solarProfileID = &parsedProfileID
	} else {
		latestProfileID, err := s.repo.GetLatestSolarProfileIDByUser(userID)
		if err != nil {
			return nil, err
		}
		solarProfileID = latestProfileID
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}

	d := &Device{
		ID:             uuid.New(),
		UserID:         userID,
		SolarProfileID: solarProfileID,
		Name:           name,
		ExternalID:     externalID,
		APIKeyPrefix:   keyPrefix(apiKey),
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := s.repo.CreateDevice(d, hashAPIKey(apiKey)); err != nil {
		return nil, err
	}

	return &CreateDeviceResponse{Device: d, APIKey: apiKey}, nil
}

// ListDevices returns all devices that belong to one user.
func (s *service) ListDevices(userID uuid.UUID) ([]*Device, error) {
	return s.repo.ListDevicesByUser(userID)
}

// GetHeartbeatSummary returns quick status for dashboard device section.
func (s *service) GetHeartbeatSummary(userID uuid.UUID) (*HeartbeatSummary, error) {
	connectedSince := time.Now().UTC().Add(-24 * time.Hour)
	return s.repo.GetHeartbeatSummaryByUser(userID, connectedSince)
}

// UpdateDevice validates and updates one user-owned device.
func (s *service) UpdateDevice(userID uuid.UUID, deviceID uuid.UUID, req UpdateDeviceRequest) (*Device, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	externalID := strings.TrimSpace(req.ExternalID)
	if externalID == "" {
		return nil, fmt.Errorf("external_id is required")
	}

	existing, err := s.repo.GetDeviceByIDForUser(deviceID, userID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return nil, fmt.Errorf("device not found")
		}
		return nil, err
	}

	var solarProfileID *uuid.UUID
	if req.SolarProfileID != nil && strings.TrimSpace(*req.SolarProfileID) != "" {
		parsedProfileID, err := uuid.Parse(strings.TrimSpace(*req.SolarProfileID))
		if err != nil {
			return nil, fmt.Errorf("solar_profile_id is invalid")
		}

		owned, err := s.repo.IsSolarProfileOwnedByUser(userID, parsedProfileID)
		if err != nil {
			return nil, err
		}
		if !owned {
			return nil, fmt.Errorf("solar_profile_id does not belong to this user")
		}

		solarProfileID = &parsedProfileID
	} else {
		solarProfileID = nil
	}

	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	existing.Name = name
	existing.ExternalID = externalID
	existing.SolarProfileID = solarProfileID
	existing.IsActive = isActive
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateDeviceForUser(existing); err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return nil, fmt.Errorf("device not found")
		}
		return nil, err
	}

	return existing, nil
}

// DeleteDevice removes one user-owned device.
func (s *service) DeleteDevice(userID uuid.UUID, deviceID uuid.UUID) error {
	err := s.repo.DeleteDeviceForUser(deviceID, userID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return fmt.Errorf("device not found")
		}
		return err
	}

	return nil
}

// RotateDeviceKey generates a fresh api key for one user-owned device.
func (s *service) RotateDeviceKey(userID uuid.UUID, deviceID uuid.UUID) (*RotateKeyResponse, error) {
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}

	if err := s.repo.RotateDeviceKey(deviceID, userID, hashAPIKey(apiKey), keyPrefix(apiKey)); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, err
	}

	return &RotateKeyResponse{APIKey: apiKey}, nil
}

// IngestTelemetry validates api key and stores telemetry with daily aggregation.
func (s *service) IngestTelemetry(apiKey string, req IngestTelemetryRequest) (*IngestTelemetryResponse, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("missing device api key")
	}

	if req.EnergyKwh <= 0 {
		return nil, fmt.Errorf("energy_kwh must be greater than 0")
	}

	eventTime, err := time.Parse(time.RFC3339, strings.TrimSpace(req.Timestamp))
	if err != nil {
		return nil, fmt.Errorf("timestamp must use RFC3339")
	}
	eventTime = eventTime.UTC()

	d, err := s.repo.FindActiveDeviceByAPIKeyHash(hashAPIKey(apiKey))
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return nil, fmt.Errorf("invalid or inactive device key")
		}
		return nil, err
	}

	providedDeviceID := strings.TrimSpace(req.DeviceID)
	if providedDeviceID != "" {
		if providedDeviceID != d.ExternalID && providedDeviceID != d.ID.String() {
			return nil, fmt.Errorf("device_id does not match authenticated device")
		}
	}

	raw := &TelemetryRaw{
		ID:          uuid.New(),
		DeviceID:    d.ID,
		UserID:      d.UserID,
		EventTime:   eventTime,
		BucketStart: bucketStart12h(eventTime),
		EnergyKwh:   req.EnergyKwh,
		PowerW:      req.PowerW,
		Lat:         req.Lat,
		Lng:         req.Lng,
		CreatedAt:   time.Now().UTC(),
	}

	inserted, err := s.repo.SaveTelemetryRaw(raw)
	if err != nil {
		return nil, err
	}

	day := normalizeDate(eventTime)
	if err := s.repo.UpdateDeviceHeartbeat(d.ID, eventTime); err != nil {
		return nil, err
	}

	if err := s.repo.RebuildActualDailyFromTelemetry(d.UserID, d.SolarProfileID, day); err != nil {
		return nil, err
	}

	message := "telemetry bucket accepted"
	if !inserted {
		message = "telemetry bucket updated"
	}

	return &IngestTelemetryResponse{
		Accepted:   true,
		Duplicate:  !inserted,
		Message:    message,
		ActualDate: day.Format(time.DateOnly),
	}, nil
}

// generateAPIKey creates a strong random key for one field device.
func generateAPIKey() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return "dvc_" + hex.EncodeToString(buf), nil
}

// hashAPIKey converts plaintext api key to deterministic SHA-256 hex string.
func hashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:])
}

// keyPrefix returns a short safe prefix for UI display.
func keyPrefix(apiKey string) string {
	if len(apiKey) < 10 {
		return apiKey
	}
	return apiKey[:10]
}

// bucketStart12h normalizes event timestamps into 00:00 or 12:00 UTC buckets.
func bucketStart12h(ts time.Time) time.Time {
	utc := ts.UTC()
	hour := 0
	if utc.Hour() >= 12 {
		hour = 12
	}
	return time.Date(utc.Year(), utc.Month(), utc.Day(), hour, 0, 0, 0, time.UTC)
}

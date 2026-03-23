package device

import (
	"time"

	"github.com/google/uuid"
)

// Device stores one user-owned field device used for telemetry ingestion.
type Device struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	SolarProfileID  *uuid.UUID `json:"solar_profile_id,omitempty"`
	Name            string     `json:"name"`
	ExternalID      string     `json:"external_id"`
	APIKeyPrefix    string     `json:"api_key_prefix"`
	IsActive        bool       `json:"is_active"`
	LastSeenAt      *time.Time `json:"last_seen_at,omitempty"`
	LastTelemetryAt *time.Time `json:"last_telemetry_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TelemetryRaw keeps the original payload sent by the device.
type TelemetryRaw struct {
	ID          uuid.UUID `json:"id"`
	DeviceID    uuid.UUID `json:"device_id"`
	UserID      uuid.UUID `json:"user_id"`
	EventTime   time.Time `json:"event_time"`
	BucketStart time.Time `json:"bucket_start"`
	EnergyKwh   float64   `json:"energy_kwh"`
	PowerW      *float64  `json:"power_w,omitempty"`
	Lat         *float64  `json:"lat,omitempty"`
	Lng         *float64  `json:"lng,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// HeartbeatSummary provides quick status for dashboard device connectivity.
type HeartbeatSummary struct {
	HasDevices        bool       `json:"has_devices"`
	TotalDevices      int        `json:"total_devices"`
	ActiveDevices     int        `json:"active_devices"`
	ConnectedDevices  int        `json:"connected_devices"`
	LatestSeenAt      *time.Time `json:"latest_seen_at,omitempty"`
	HeartbeatWindowHr int        `json:"heartbeat_window_hr"`
}

// CreateDeviceRequest is used by authenticated users to register a new device.
type CreateDeviceRequest struct {
	Name           string  `json:"name"`
	ExternalID     string  `json:"external_id"`
	SolarProfileID *string `json:"solar_profile_id,omitempty"`
	PlanTier       string  `json:"-"`
}

// UpdateDeviceRequest is used by authenticated users to update one registered device.
type UpdateDeviceRequest struct {
	Name           string  `json:"name"`
	ExternalID     string  `json:"external_id"`
	SolarProfileID *string `json:"solar_profile_id,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// IngestTelemetryRequest is the payload posted by a field device.
type IngestTelemetryRequest struct {
	DeviceID  string   `json:"device_id"`
	Timestamp string   `json:"timestamp"`
	EnergyKwh float64  `json:"energy_kwh"`
	PowerW    *float64 `json:"power_w,omitempty"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
}

// CreateDeviceResponse returns metadata and plaintext api key once.
type CreateDeviceResponse struct {
	Device *Device `json:"device"`
	APIKey string  `json:"api_key"`
}

// RotateKeyResponse returns refreshed plaintext api key once.
type RotateKeyResponse struct {
	APIKey string `json:"api_key"`
}

// IngestTelemetryResponse reports ingestion status for one device payload.
type IngestTelemetryResponse struct {
	Accepted   bool   `json:"accepted"`
	Duplicate  bool   `json:"duplicate"`
	Message    string `json:"message"`
	ActualDate string `json:"actual_date"`
}

package device

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository defines persistence operations for device integration.
type Repository interface {
	CreateDevice(d *Device, apiKeyHash string) error
	ListDevicesByUser(userID uuid.UUID) ([]*Device, error)
	GetDeviceByIDForUser(deviceID uuid.UUID, userID uuid.UUID) (*Device, error)
	UpdateDeviceForUser(d *Device) error
	DeleteDeviceForUser(deviceID uuid.UUID, userID uuid.UUID) error
	IsSolarProfileOwnedByUser(userID uuid.UUID, solarProfileID uuid.UUID) (bool, error)
	GetLatestSolarProfileIDByUser(userID uuid.UUID) (*uuid.UUID, error)
	RotateDeviceKey(deviceID uuid.UUID, userID uuid.UUID, apiKeyHash string, apiKeyPrefix string) error
	GetHeartbeatSummaryByUser(userID uuid.UUID, connectedSince time.Time) (*HeartbeatSummary, error)
	FindActiveDeviceByAPIKeyHash(apiKeyHash string) (*Device, error)
	SaveTelemetryRaw(t *TelemetryRaw) (bool, error)
	RebuildActualDailyFromTelemetry(userID uuid.UUID, solarProfileID *uuid.UUID, day time.Time) error
	UpdateDeviceHeartbeat(deviceID uuid.UUID, seenAt time.Time) error
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new device repository.
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// CreateDevice inserts one user-owned device with api key hash.
func (r *repository) CreateDevice(d *Device, apiKeyHash string) error {
	query := `
		INSERT INTO devices (id, user_id, solar_profile_id, name, external_id, api_key_hash, api_key_prefix, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query, d.ID, d.UserID, d.SolarProfileID, d.Name, d.ExternalID, apiKeyHash, d.APIKeyPrefix, d.IsActive, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create device: %w", err)
	}
	return nil
}

// ListDevicesByUser returns all devices for one authenticated user.
func (r *repository) ListDevicesByUser(userID uuid.UUID) ([]*Device, error) {
	query := `
		SELECT id, user_id, solar_profile_id, name, external_id, api_key_prefix, is_active, last_seen_at, last_telemetry_at, created_at, updated_at
		FROM devices
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("list devices by user: %w", err)
	}
	defer rows.Close()

	devices := []*Device{}
	for rows.Next() {
		d := &Device{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.SolarProfileID, &d.Name, &d.ExternalID, &d.APIKeyPrefix, &d.IsActive, &d.LastSeenAt, &d.LastTelemetryAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device row: %w", err)
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate devices rows: %w", err)
	}

	return devices, nil
}

// GetDeviceByIDForUser returns one device only if it belongs to the user.
func (r *repository) GetDeviceByIDForUser(deviceID uuid.UUID, userID uuid.UUID) (*Device, error) {
	query := `
		SELECT id, user_id, solar_profile_id, name, external_id, api_key_prefix, is_active, last_seen_at, last_telemetry_at, created_at, updated_at
		FROM devices
		WHERE id = $1 AND user_id = $2
	`

	d := &Device{}
	if err := r.db.QueryRow(query, deviceID, userID).Scan(&d.ID, &d.UserID, &d.SolarProfileID, &d.Name, &d.ExternalID, &d.APIKeyPrefix, &d.IsActive, &d.LastSeenAt, &d.LastTelemetryAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get device by id for user: %w", err)
	}
	return d, nil
}

// UpdateDeviceForUser updates mutable fields for one user-owned device.
func (r *repository) UpdateDeviceForUser(d *Device) error {
	query := `
		UPDATE devices
		SET solar_profile_id = $3,
		    name = $4,
		    external_id = $5,
		    is_active = $6,
		    updated_at = $7
		WHERE id = $1 AND user_id = $2
	`

	res, err := r.db.Exec(query, d.ID, d.UserID, d.SolarProfileID, d.Name, d.ExternalID, d.IsActive, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update device: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected update device: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteDeviceForUser deletes one device owned by one user.
func (r *repository) DeleteDeviceForUser(deviceID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM devices WHERE id = $1 AND user_id = $2`

	res, err := r.db.Exec(query, deviceID, userID)
	if err != nil {
		return fmt.Errorf("delete device: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected delete device: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// IsSolarProfileOwnedByUser checks whether one solar profile belongs to one user.
func (r *repository) IsSolarProfileOwnedByUser(userID uuid.UUID, solarProfileID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM solar_profiles WHERE id = $1 AND user_id = $2)`
	var exists bool
	if err := r.db.QueryRow(query, solarProfileID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check solar profile ownership: %w", err)
	}
	return exists, nil
}

// GetLatestSolarProfileIDByUser returns the newest solar profile id for one user.
func (r *repository) GetLatestSolarProfileIDByUser(userID uuid.UUID) (*uuid.UUID, error) {
	query := `SELECT id FROM solar_profiles WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	var id uuid.UUID
	err := r.db.QueryRow(query, userID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest solar profile id by user: %w", err)
	}
	return &id, nil
}

// GetHeartbeatSummaryByUser returns quick connectivity summary for one user.
func (r *repository) GetHeartbeatSummaryByUser(userID uuid.UUID, connectedSince time.Time) (*HeartbeatSummary, error) {
	query := `
		SELECT
			COUNT(*) AS total_devices,
			COUNT(*) FILTER (WHERE is_active = TRUE) AS active_devices,
			COUNT(*) FILTER (WHERE is_active = TRUE AND last_seen_at IS NOT NULL AND last_seen_at >= $2) AS connected_devices,
			MAX(last_seen_at) AS latest_seen_at
		FROM devices
		WHERE user_id = $1
	`

	s := &HeartbeatSummary{HeartbeatWindowHr: 24}
	if err := r.db.QueryRow(query, userID, connectedSince).Scan(&s.TotalDevices, &s.ActiveDevices, &s.ConnectedDevices, &s.LatestSeenAt); err != nil {
		return nil, fmt.Errorf("get heartbeat summary by user: %w", err)
	}
	s.HasDevices = s.TotalDevices > 0
	return s, nil
}

// RotateDeviceKey updates stored key hash for one user-owned device.
func (r *repository) RotateDeviceKey(deviceID uuid.UUID, userID uuid.UUID, apiKeyHash string, apiKeyPrefix string) error {
	query := `
		UPDATE devices
		SET api_key_hash = $3,
		    api_key_prefix = $4,
		    updated_at = $5
		WHERE id = $1 AND user_id = $2
	`

	res, err := r.db.Exec(query, deviceID, userID, apiKeyHash, apiKeyPrefix, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("rotate device key: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected rotate device key: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// FindActiveDeviceByAPIKeyHash returns the active device associated to one api key hash.
func (r *repository) FindActiveDeviceByAPIKeyHash(apiKeyHash string) (*Device, error) {
	query := `
		SELECT id, user_id, solar_profile_id, name, external_id, api_key_prefix, is_active, last_seen_at, last_telemetry_at, created_at, updated_at
		FROM devices
		WHERE api_key_hash = $1 AND is_active = TRUE
	`

	d := &Device{}
	if err := r.db.QueryRow(query, apiKeyHash).Scan(&d.ID, &d.UserID, &d.SolarProfileID, &d.Name, &d.ExternalID, &d.APIKeyPrefix, &d.IsActive, &d.LastSeenAt, &d.LastTelemetryAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, fmt.Errorf("find active device by api key hash: %w", err)
	}
	return d, nil
}

// SaveTelemetryRaw upserts one telemetry point per device per 12-hour bucket.
func (r *repository) SaveTelemetryRaw(t *TelemetryRaw) (bool, error) {
	query := `
		INSERT INTO telemetry_raw (id, device_id, user_id, event_time, bucket_start, energy_kwh, power_w, lat, lng, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (device_id, bucket_start) DO UPDATE
		SET event_time = EXCLUDED.event_time,
		    energy_kwh = EXCLUDED.energy_kwh,
		    power_w = EXCLUDED.power_w,
		    lat = EXCLUDED.lat,
		    lng = EXCLUDED.lng
		RETURNING (xmax = 0) AS inserted
	`

	var inserted bool
	err := r.db.QueryRow(query, t.ID, t.DeviceID, t.UserID, t.EventTime, t.BucketStart, t.EnergyKwh, t.PowerW, t.Lat, t.Lng, t.CreatedAt).Scan(&inserted)
	if err != nil {
		return false, fmt.Errorf("save telemetry raw: %w", err)
	}

	return inserted, nil
}

// RebuildActualDailyFromTelemetry recomputes one day actual from telemetry snapshots.
func (r *repository) RebuildActualDailyFromTelemetry(userID uuid.UUID, solarProfileID *uuid.UUID, day time.Time) error {
	query := `
		WITH total AS (
			SELECT COALESCE(SUM(energy_kwh), 0) AS total_kwh
			FROM telemetry_raw tr
			JOIN devices d ON d.id = tr.device_id
			WHERE tr.user_id = $1
			  AND d.solar_profile_id IS NOT DISTINCT FROM $2
			  AND (tr.event_time AT TIME ZONE 'UTC')::date = $3
		)
		INSERT INTO actual_daily (id, user_id, solar_profile_id, date, actual_kwh, source, created_at)
		VALUES ($4, $1, $2, $3, (SELECT total_kwh FROM total), 'iot', $5)
		ON CONFLICT (user_id, solar_profile_id, date) DO UPDATE
		SET actual_kwh = (SELECT total_kwh FROM total),
		    source = 'iot'
	`

	_, err := r.db.Exec(query, userID, solarProfileID, normalizeDate(day), uuid.New(), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("rebuild actual daily from telemetry: %w", err)
	}
	return nil
}

// UpdateDeviceHeartbeat stores the latest device seen timestamp.
func (r *repository) UpdateDeviceHeartbeat(deviceID uuid.UUID, seenAt time.Time) error {
	query := `
		UPDATE devices
		SET last_seen_at = $2,
		    last_telemetry_at = $2,
		    updated_at = $3
		WHERE id = $1
	`

	res, err := r.db.Exec(query, deviceID, seenAt.UTC(), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update device heartbeat: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected update device heartbeat: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// normalizeDate strips time component for date columns.
func normalizeDate(input time.Time) time.Time {
	return time.Date(input.Year(), input.Month(), input.Day(), 0, 0, 0, 0, time.UTC)
}

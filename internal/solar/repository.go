package solar

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Repository defines data access methods for solar profiles
type Repository interface {
	CreateSolarProfile(p *SolarProfile) (*SolarProfile, error)
	UpdateSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID, p *SolarProfile) (*SolarProfile, error)
	DeleteSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID) error
	GetSolarProfilesByUserID(userID uuid.UUID) ([]*SolarProfile, error)
	GetSolarProfileByUserID(userID uuid.UUID) (*SolarProfile, error)
	GetSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID) (*SolarProfile, error)
	GetAllSolarProfiles() ([]*SolarProfile, error)
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new solar profile repository
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// CreateSolarProfile inserts a solar profile for a user.
func (r *repository) CreateSolarProfile(p *SolarProfile) (*SolarProfile, error) {
	query := `
		INSERT INTO solar_profiles (id, user_id, site_name, capacity_kwp, lat, lng, tilt, azimuth, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, site_name, capacity_kwp, lat, lng, tilt, azimuth, created_at
	`

	stored := &SolarProfile{}
	err := r.db.QueryRow(query, p.ID, p.UserID, p.SiteName, p.CapacityKwp, p.Lat, p.Lng, p.Tilt, p.Azimuth, p.CreatedAt).Scan(
		&stored.ID,
		&stored.UserID,
		&stored.SiteName,
		&stored.CapacityKwp,
		&stored.Lat,
		&stored.Lng,
		&stored.Tilt,
		&stored.Azimuth,
		&stored.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create solar profile: %w", err)
	}
	return stored, nil
}

// UpdateSolarProfileByIDAndUserID updates one solar profile owned by one user.
func (r *repository) UpdateSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID, p *SolarProfile) (*SolarProfile, error) {
	query := `
		UPDATE solar_profiles
		SET site_name = $3,
		    capacity_kwp = $4,
		    lat = $5,
		    lng = $6,
		    tilt = $7,
		    azimuth = $8
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, site_name, capacity_kwp, lat, lng, tilt, azimuth, created_at
	`

	stored := &SolarProfile{}
	err := r.db.QueryRow(query, profileID, userID, p.SiteName, p.CapacityKwp, p.Lat, p.Lng, p.Tilt, p.Azimuth).Scan(
		&stored.ID,
		&stored.UserID,
		&stored.SiteName,
		&stored.CapacityKwp,
		&stored.Lat,
		&stored.Lng,
		&stored.Tilt,
		&stored.Azimuth,
		&stored.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update solar profile: %w", err)
	}

	return stored, nil
}

// DeleteSolarProfileByIDAndUserID deletes one solar profile owned by one user.
func (r *repository) DeleteSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM solar_profiles WHERE id = $1 AND user_id = $2`

	res, err := r.db.Exec(query, profileID, userID)
	if err != nil {
		return fmt.Errorf("delete solar profile: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected delete solar profile: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetSolarProfilesByUserID returns all solar profiles belonging to one user.
func (r *repository) GetSolarProfilesByUserID(userID uuid.UUID) ([]*SolarProfile, error) {
	query := `
		SELECT id, user_id, site_name, capacity_kwp, lat, lng, tilt, azimuth, created_at
		FROM solar_profiles
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("get solar profiles by user id: %w", err)
	}
	defer rows.Close()

	profiles := []*SolarProfile{}
	for rows.Next() {
		p := &SolarProfile{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.SiteName, &p.CapacityKwp, &p.Lat, &p.Lng, &p.Tilt, &p.Azimuth, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan solar profile by user id: %w", err)
		}
		profiles = append(profiles, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate solar profiles by user id: %w", err)
	}

	return profiles, nil
}

// GetSolarProfileByUserID fetches the solar profile belonging to a user
func (r *repository) GetSolarProfileByUserID(userID uuid.UUID) (*SolarProfile, error) {
	query := `
		SELECT id, user_id, site_name, capacity_kwp, lat, lng, tilt, azimuth, created_at
		FROM solar_profiles WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRow(query, userID)

	p := &SolarProfile{}
	if err := row.Scan(&p.ID, &p.UserID, &p.SiteName, &p.CapacityKwp, &p.Lat, &p.Lng, &p.Tilt, &p.Azimuth, &p.CreatedAt); err != nil {
		return nil, fmt.Errorf("get solar profile by user id: %w", err)
	}
	return p, nil
}

// GetSolarProfileByIDAndUserID fetches one solar profile by id scoped to one user.
func (r *repository) GetSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID) (*SolarProfile, error) {
	query := `
		SELECT id, user_id, site_name, capacity_kwp, lat, lng, tilt, azimuth, created_at
		FROM solar_profiles
		WHERE id = $1 AND user_id = $2
	`

	p := &SolarProfile{}
	if err := r.db.QueryRow(query, profileID, userID).Scan(&p.ID, &p.UserID, &p.SiteName, &p.CapacityKwp, &p.Lat, &p.Lng, &p.Tilt, &p.Azimuth, &p.CreatedAt); err != nil {
		return nil, fmt.Errorf("get solar profile by id and user id: %w", err)
	}

	return p, nil
}

// GetAllSolarProfiles returns every solar profile in the database
func (r *repository) GetAllSolarProfiles() ([]*SolarProfile, error) {
	query := `
		SELECT id, user_id, site_name, capacity_kwp, lat, lng, tilt, azimuth, created_at
		FROM solar_profiles ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all solar profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*SolarProfile
	for rows.Next() {
		p := &SolarProfile{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.SiteName, &p.CapacityKwp, &p.Lat, &p.Lng, &p.Tilt, &p.Azimuth, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan solar profile: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

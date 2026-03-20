package solar

import (
	"time"

	"github.com/google/uuid"
)

// SolarProfile represents a user's solar panel installation
type SolarProfile struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	SiteName    string    `json:"site_name"`
	CapacityKwp float64   `json:"capacity_kwp"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Tilt        *float64  `json:"tilt,omitempty"`
	Azimuth     *float64  `json:"azimuth,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateSolarProfileRequest holds data needed to create a solar panel profile
type CreateSolarProfileRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	SiteName    string    `json:"site_name"`
	CapacityKwp float64   `json:"capacity_kwp"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Tilt        *float64  `json:"tilt,omitempty"`
	Azimuth     *float64  `json:"azimuth,omitempty"`
}

// UpdateSolarProfileRequest holds data needed to update one solar panel profile.
type UpdateSolarProfileRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	SiteName    string    `json:"site_name"`
	CapacityKwp float64   `json:"capacity_kwp"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Tilt        *float64  `json:"tilt,omitempty"`
	Azimuth     *float64  `json:"azimuth,omitempty"`
}

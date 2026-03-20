package solar

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service defines business logic operations for solar profiles
type Service interface {
	CreateSolarProfile(req CreateSolarProfileRequest) (*SolarProfile, error)
	UpdateSolarProfile(profileID uuid.UUID, req UpdateSolarProfileRequest) (*SolarProfile, error)
	DeleteSolarProfile(profileID uuid.UUID, userID uuid.UUID) error
	GetSolarProfilesByUserID(userID uuid.UUID) ([]*SolarProfile, error)
	GetSolarProfileByUserID(userID uuid.UUID) (*SolarProfile, error)
	GetSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID) (*SolarProfile, error)
	GetAllSolarProfiles() ([]*SolarProfile, error)
}

type service struct {
	repo Repository
}

// NewService creates a new solar profile service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateSolarProfile validates and saves a solar profile to the database
func (s *service) CreateSolarProfile(req CreateSolarProfileRequest) (*SolarProfile, error) {
	if req.CapacityKwp <= 0 {
		return nil, fmt.Errorf("capacity_kwp must be greater than 0")
	}
	if req.Lat == 0 || req.Lng == 0 {
		return nil, fmt.Errorf("lat and lng are required")
	}

	siteName := strings.TrimSpace(req.SiteName)
	if siteName == "" {
		siteName = "Main Site"
	}

	p := &SolarProfile{
		ID:          uuid.New(),
		UserID:      req.UserID,
		SiteName:    siteName,
		CapacityKwp: req.CapacityKwp,
		Lat:         req.Lat,
		Lng:         req.Lng,
		Tilt:        req.Tilt,
		Azimuth:     req.Azimuth,
		CreatedAt:   time.Now().UTC(),
	}

	return s.repo.CreateSolarProfile(p)
}

// UpdateSolarProfile validates and updates one user-owned solar profile.
func (s *service) UpdateSolarProfile(profileID uuid.UUID, req UpdateSolarProfileRequest) (*SolarProfile, error) {
	if req.CapacityKwp <= 0 {
		return nil, fmt.Errorf("capacity_kwp must be greater than 0")
	}
	if req.Lat == 0 || req.Lng == 0 {
		return nil, fmt.Errorf("lat and lng are required")
	}

	siteName := strings.TrimSpace(req.SiteName)
	if siteName == "" {
		siteName = "Main Site"
	}

	p := &SolarProfile{
		SiteName:    siteName,
		CapacityKwp: req.CapacityKwp,
		Lat:         req.Lat,
		Lng:         req.Lng,
		Tilt:        req.Tilt,
		Azimuth:     req.Azimuth,
	}

	updated, err := s.repo.UpdateSolarProfileByIDAndUserID(profileID, req.UserID, p)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return nil, fmt.Errorf("solar profile not found")
		}
		return nil, err
	}

	return updated, nil
}

// DeleteSolarProfile removes one user-owned solar profile.
func (s *service) DeleteSolarProfile(profileID uuid.UUID, userID uuid.UUID) error {
	err := s.repo.DeleteSolarProfileByIDAndUserID(profileID, userID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return fmt.Errorf("solar profile not found")
		}
		return err
	}

	return nil
}

// GetSolarProfilesByUserID retrieves all solar profiles for a given user.
func (s *service) GetSolarProfilesByUserID(userID uuid.UUID) ([]*SolarProfile, error) {
	return s.repo.GetSolarProfilesByUserID(userID)
}

// GetSolarProfileByUserID retrieves the solar profile for a given user
func (s *service) GetSolarProfileByUserID(userID uuid.UUID) (*SolarProfile, error) {
	return s.repo.GetSolarProfileByUserID(userID)
}

// GetSolarProfileByIDAndUserID retrieves one profile by id scoped to one user.
func (s *service) GetSolarProfileByIDAndUserID(profileID uuid.UUID, userID uuid.UUID) (*SolarProfile, error) {
	return s.repo.GetSolarProfileByIDAndUserID(profileID, userID)
}

// GetAllSolarProfiles returns all solar profiles
func (s *service) GetAllSolarProfiles() ([]*SolarProfile, error) {
	return s.repo.GetAllSolarProfiles()
}

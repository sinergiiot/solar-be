package weatherbaseline

import (
	"context"
	"database/sql"
	"time"
)

// BaselineType values: "synthetic", "site", "blended"
type BaselineType string

const (
	BaselineSynthetic BaselineType = "synthetic"
	BaselineSite      BaselineType = "site"
	BaselineBlended   BaselineType = "blended"
)

type WeatherBaseline struct {
	ID            int64
	ProfileID     string
	UserID        string
	BaselineType  BaselineType
	BaselineValue float64
	SampleCount   int
	ValidFrom     time.Time
	ValidTo       time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repository interface {
	GetBaseline(ctx context.Context, profileID, userID string, baselineType BaselineType) (*WeatherBaseline, error)
	SaveBaseline(ctx context.Context, b *WeatherBaseline) error
	DB() *sql.DB
}
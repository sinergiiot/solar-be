package rec

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MwhAccumulator struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	SolarProfileID   *uuid.UUID `json:"solar_profile_id,omitempty"`
	CumulativeKwh    float64    `json:"cumulative_kwh"`
	LastUpdatedAt    time.Time  `json:"last_updated_at"`
	MilestoneReached bool       `json:"milestone_reached"`
}

type Service interface {
	UpdateAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID, kwh float64) error
	GetAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID) (*MwhAccumulator, error)
	GetTotalMwhForUser(ctx context.Context, userID uuid.UUID) (float64, error)
}

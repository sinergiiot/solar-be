package admin

import (
	"context"

	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
)

type AdminUserRow struct {
	user.User
	PlanTier string `json:"plan_tier"`
}

type SystemStats struct {
	TotalUsers      int     `json:"total_users"`
	TotalPro        int     `json:"total_pro"`
	TotalEnterprise int     `json:"total_enterprise"`
	TotalKwh        float64 `json:"total_kwh"`
	TotalProfiles   int     `json:"total_profiles"`
}

type Service interface {
	GetAllUsersWithTiers(ctx context.Context) ([]AdminUserRow, error)
	UpdateUserTier(ctx context.Context, userID uuid.UUID, newTier string) error
	GetSystemStats(ctx context.Context) (*SystemStats, error)
}

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

type Service interface {
	GetAllUsersWithTiers(ctx context.Context) ([]AdminUserRow, error)
	UpdateUserTier(ctx context.Context, userID uuid.UUID, newTier string) error
}

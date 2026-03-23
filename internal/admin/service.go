package admin

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
)

type service struct {
	db       *sql.DB
	userSvc  user.Service
}

func NewService(db *sql.DB, userSvc user.Service) Service {
	return &service{
		db:      db,
		userSvc: userSvc,
	}
}

func (s *service) GetAllUsersWithTiers(ctx context.Context) ([]AdminUserRow, error) {
	users, err := s.userSvc.GetAllUsers()
	if err != nil {
		return nil, err
	}

	query := `SELECT user_id, plan_tier FROM notification_preferences`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tiers: %w", err)
	}
	defer rows.Close()

	tierMap := make(map[uuid.UUID]string)
	for rows.Next() {
		var uid uuid.UUID
		var tier string
		if err := rows.Scan(&uid, &tier); err == nil {
			tierMap[uid] = tier
		}
	}

	result := make([]AdminUserRow, len(users))
	for i, u := range users {
		tier := "free"
		if val, ok := tierMap[u.ID]; ok {
			tier = val
		}
		result[i] = AdminUserRow{
			User:     *u,
			PlanTier: tier,
		}
	}

	return result, nil
}

func (s *service) UpdateUserTier(ctx context.Context, userID uuid.UUID, newTier string) error {
	query := `
		INSERT INTO notification_preferences (user_id, plan_tier)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET plan_tier = EXCLUDED.plan_tier, updated_at = NOW()
	`
	_, err := s.db.ExecContext(ctx, query, userID, newTier)
	if err != nil {
		return fmt.Errorf("update user tier: %w", err)
	}
	return nil
}

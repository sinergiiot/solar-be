package rec

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	UpsertAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID, kwh float64) (float64, error)
	SetMilestoneReached(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID) error
	GetAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID) (*MwhAccumulator, error)
	GetTotalKwhForUser(ctx context.Context, userID uuid.UUID) (float64, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) UpsertAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID, kwh float64) (float64, error) {
	query := `
		INSERT INTO mwh_accumulators (user_id, solar_profile_id, cumulative_kwh, last_updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, solar_profile_id)
		DO UPDATE SET 
			cumulative_kwh = mwh_accumulators.cumulative_kwh + EXCLUDED.cumulative_kwh,
			last_updated_at = EXCLUDED.last_updated_at
		RETURNING cumulative_kwh
	`
	var newTotal float64
	err := r.db.QueryRowContext(ctx, query, userID, profileID, kwh, time.Now().UTC()).Scan(&newTotal)
	return newTotal, err
}

func (r *repository) SetMilestoneReached(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID) error {
	query := `
		UPDATE mwh_accumulators
		SET milestone_reached = true
		WHERE user_id = $1 AND solar_profile_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, userID, profileID)
	return err
}

func (r *repository) GetAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID) (*MwhAccumulator, error) {
	query := `
		SELECT id, user_id, solar_profile_id, cumulative_kwh, last_updated_at, milestone_reached
		FROM mwh_accumulators
		WHERE user_id = $1 AND solar_profile_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, userID, profileID)
	
	var acc MwhAccumulator
	err := row.Scan(&acc.ID, &acc.UserID, &acc.SolarProfileID, &acc.CumulativeKwh, &acc.LastUpdatedAt, &acc.MilestoneReached)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *repository) GetTotalKwhForUser(ctx context.Context, userID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(cumulative_kwh), 0)
		FROM mwh_accumulators
		WHERE user_id = $1
	`
	var total float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}

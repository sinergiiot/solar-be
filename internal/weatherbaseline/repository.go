package weatherbaseline

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository implements WeatherBaseline DB access

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetBaseline(ctx context.Context, profileID, userID string, baselineType BaselineType) (*WeatherBaseline, error) {
	query := `SELECT id, profile_id, user_id, baseline_type, baseline_value, sample_count, valid_from, valid_to, created_at, updated_at
		FROM weather_baselines WHERE profile_id = $1 AND user_id = $2 AND baseline_type = $3`
	row := r.db.QueryRowContext(ctx, query, profileID, userID, string(baselineType))
	b := &WeatherBaseline{}
	if err := row.Scan(&b.ID, &b.ProfileID, &b.UserID, &b.BaselineType, &b.BaselineValue, &b.SampleCount, &b.ValidFrom, &b.ValidTo, &b.CreatedAt, &b.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get baseline: %w", err)
	}
	return b, nil
}

func (r *repository) SaveBaseline(ctx context.Context, b *WeatherBaseline) error {
	query := `INSERT INTO weather_baselines (profile_id, user_id, baseline_type, baseline_value, sample_count, valid_from, valid_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
		ON CONFLICT (profile_id, user_id, baseline_type) DO UPDATE
		SET baseline_value = EXCLUDED.baseline_value,
		    sample_count = EXCLUDED.sample_count,
		    valid_from = EXCLUDED.valid_from,
		    valid_to = EXCLUDED.valid_to,
		    updated_at = now()`
	_, err := r.db.ExecContext(ctx, query, b.ProfileID, b.UserID, string(b.BaselineType), b.BaselineValue, b.SampleCount, b.ValidFrom, b.ValidTo)
	if err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}
	return nil
}

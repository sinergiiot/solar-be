package forecast

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/akbarsenawijaya/solar-forecast/internal/weatherbaseline"
	"github.com/google/uuid"
)

// DeltaWFResult holds the result of delta weather factor calculation
// baselineType: 'synthetic' | 'site' | 'blended'
type DeltaWFResult struct {
	DeltaWF      float64
	BaselineType string
}

// ComputeDeltaWF calculates delta weather factor for a given profile/user/cloudCoverToday
// countActualDaysRepo is a minimal interface for counting valid actual days
type countActualDaysRepo interface {
	CountValidActualDays(ctx context.Context, userID uuid.UUID, profileID uuid.UUID) (int, error)
}

// ComputeDeltaWF calculates delta weather factor for a given profile/user/cloudCoverToday
// Accepts repo to count valid actual days (actual_kwh > 0)
func ComputeDeltaWF(ctx context.Context, repo countActualDaysRepo, wbSvc weatherbaseline.Service, profileID, userID string, cloudCoverToday float64, lat, lng float64) (DeltaWFResult, error) {
	const NThreshold = 14
	// 1. Count n_actual (valid actual days)
	nActual, err := countActualDays(ctx, repo, profileID, userID)
	if err != nil {
		return DeltaWFResult{}, fmt.Errorf("count actual days: %w", err)
	}
	var baseline float64
	var baselineType string
	if nActual == 0 {
		// Cold start: use synthetic
		baseline, _, err = wbSvc.GetSyntheticBaseline(ctx, profileID, userID, lat, lng)
		baselineType = "synthetic"
	} else if nActual >= NThreshold {
		// Calibrated: use site
		baseline, _, err = wbSvc.GetSiteBaseline(ctx, profileID, userID)
		baselineType = "site"
	} else {
		// Transition: blend
		w := float64(nActual) / float64(NThreshold)
		cold, _, err1 := wbSvc.GetSyntheticBaseline(ctx, profileID, userID, lat, lng)
		site, _, err2 := wbSvc.GetSiteBaseline(ctx, profileID, userID)
		if err1 != nil && err2 != nil {
			return DeltaWFResult{}, errors.New("no baseline available")
		}
		if err1 != nil {
			cold = site
		}
		if err2 != nil {
			site = cold
		}
		baseline = (1-w)*cold + w*site
		baselineType = "blended"
	}
	if err != nil {
		return DeltaWFResult{}, fmt.Errorf("get baseline: %w", err)
	}
	if baseline >= 95 {
		// fallback to absolute
		return DeltaWFResult{
			DeltaWF: 1 - cloudCoverToday/100,
			BaselineType: baselineType,
		}, nil
	}
	deltaWF := (1 - cloudCoverToday/100) / (1 - baseline/100)
	deltaWF = clamp(deltaWF, 0.5, 1.5)
	return DeltaWFResult{
		DeltaWF:      deltaWF,
		BaselineType: baselineType,
	}, nil
}

func clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}

// countActualDays queries actual_daily for valid days (actual_kwh > 0) for a profile/user
func countActualDays(ctx context.Context, repo countActualDaysRepo, profileID, userID string) (int, error) {
	// Parse UUIDs
	profileUUID, err := parseUUID(profileID)
	if err != nil {
		return 0, fmt.Errorf("invalid profileID: %w", err)
	}
	userUUID, err := parseUUID(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid userID: %w", err)
	}
	return repo.CountValidActualDays(ctx, userUUID, profileUUID)
}

// parseUUID parses a string to uuid.UUID, returns error if invalid
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

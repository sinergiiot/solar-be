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
	WeatherFactor float64 // actual (1 - cloudCover/100)
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
		// Cold start: use synthetic baseline
		b, _, err := wbSvc.GetSyntheticBaseline(ctx, profileID, userID, lat, lng)
		if err != nil {
			return DeltaWFResult{}, fmt.Errorf("get synthetic baseline: %w", err)
		}
		baseline = b
		baselineType = "synthetic"

	} else if nActual >= NThreshold {
		// Calibrated: use site baseline
		b, _, err := wbSvc.GetSiteBaseline(ctx, profileID, userID)
		if err != nil {
			return DeltaWFResult{}, fmt.Errorf("get site baseline: %w", err)
		}
		baseline = b
		baselineType = "site"

	} else {
		// Transition: blend synthetic and site baseline
		// w = 0.0 when n=0 (full synthetic), w = 1.0 when n=NThreshold (full site)
		w := float64(nActual) / float64(NThreshold)
		cold, _, err1 := wbSvc.GetSyntheticBaseline(ctx, profileID, userID, lat, lng)
		site, _, err2 := wbSvc.GetSiteBaseline(ctx, profileID, userID)
		if err1 != nil && err2 != nil {
			return DeltaWFResult{}, errors.New("no baseline available for transition mode")
		}
		if err1 != nil {
			cold = site
		} // fallback: use site only
		if err2 != nil {
			site = cold
		} // fallback: use synthetic only
		baseline = (1-w)*cold + w*site
		baselineType = "blended"
	}
	       if baseline >= 95 {
		       // fallback to absolute
		       wf := 1 - cloudCoverToday/100
		       return DeltaWFResult{
			       DeltaWF: wf,
			       WeatherFactor: wf,
			       BaselineType: baselineType,
		       }, nil
	       }
	       wf := 1 - cloudCoverToday/100
	       deltaWF := wf / (1 - baseline/100)
	       // Clamp deltaWF tighter: max 1.1 (was 1.5)
	       deltaWF = clamp(deltaWF, 0.5, 1.1)
	       return DeltaWFResult{
		       DeltaWF:      deltaWF,
		       WeatherFactor: wf,
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

// DetermineWeatherRisk returns weather risk status based on cloud cover and deltaWF.
// cloudCover : 0–100 (integer, from WeatherDaily.CloudCover)
// deltaWF    : 0.5–1.1 (float64, after clamp, from DeltaWFResult.DeltaWF)
// returns    : "Risiko Tinggi" | "Risiko Sedang" | "Risiko Rendah"
func DetermineWeatherRisk(cloudCover int, deltaWF float64) string {
	// 1. Risiko Tinggi : cloud_cover > 90  OR  delta_wf < 0.60
	if cloudCover > 90 || deltaWF < 0.60 {
		return "Potensi Drop Drastis"
	}
	// 2. Risiko Sedang : cloud_cover 80–90 OR  delta_wf 0.60–0.85
	if cloudCover >= 80 || deltaWF <= 0.85 {
		return "Potensi Fluktuasi"
	}
	// 3. Risiko Rendah
	return "Produksi Optimal"
}

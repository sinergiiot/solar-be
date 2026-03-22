package weatherbaseline

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Service provides baseline calculation and caching

type Service interface {
	GetSyntheticBaseline(ctx context.Context, profileID, userID string, lat, lng float64) (float64, int, error)
	GetSiteBaseline(ctx context.Context, profileID, userID string) (float64, int, error)
}

type service struct {
	repo    Repository
	baseURL string
}

func NewService(repo Repository, baseURL string) Service {
	return &service{repo: repo, baseURL: baseURL}
}

// GetSyntheticBaseline fetches/caches 30-day Open-Meteo baseline
func (s *service) GetSyntheticBaseline(ctx context.Context, profileID, userID string, lat, lng float64) (float64, int, error) {
	// 1. Check cache
	b, err := s.repo.GetBaseline(ctx, profileID, userID, BaselineSynthetic)
	if err == nil && b != nil && b.ValidTo.After(time.Now()) {
		return b.BaselineValue, b.SampleCount, nil
	}
	// 2. Fetch 30-day historical from Open-Meteo
	dateTo := time.Now().AddDate(0, 0, -1)
	dateFrom := dateTo.AddDate(0, 0, -30)
	url := fmt.Sprintf("https://archive-api.open-meteo.com/v1/archive?latitude=%.4f&longitude=%.4f&start_date=%s&end_date=%s&daily=cloud_cover_mean&timezone=Asia%%2FJakarta", lat, lng, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("fetch open-meteo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("open-meteo status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("read open-meteo: %w", err)
	}
	var api struct {
		Daily struct {
			CloudCoverMean []float64 `json:"cloud_cover_mean"`
		} `json:"daily"`
	}
	if err := json.Unmarshal(body, &api); err != nil {
		return 0, 0, fmt.Errorf("unmarshal open-meteo: %w", err)
	}
	var sum float64
	var count int
	for _, v := range api.Daily.CloudCoverMean {
		if v >= 0 && v <= 100 {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0, 0, errors.New("no valid cloud_cover_mean data")
	}
	avg := sum / float64(count)
	// 3. Cache result
	baseline := &WeatherBaseline{
		ProfileID:     profileID,
		UserID:        userID,
		BaselineType:  BaselineSynthetic,
		BaselineValue: avg,
		SampleCount:   count,
		ValidFrom:     dateFrom,
		ValidTo:       time.Now().AddDate(0, 0, 1), // Cache valid for 24 hours (until tomorrow)
	}
	_ = s.repo.SaveBaseline(ctx, baseline)
	return avg, count, nil
}


// GetSiteBaseline computes baseline from site actuals (valid days with actual > 0)
func (s *service) GetSiteBaseline(ctx context.Context, profileID, userID string) (float64, int, error) {
	// 1. Check cache
	b, err := s.repo.GetBaseline(ctx, profileID, userID, BaselineSite)
	if err == nil && b != nil && b.ValidTo.After(time.Now()) {
		return b.BaselineValue, b.SampleCount, nil
	}
	// 2. Query actuals join weather_daily for valid days
	// (Assume db available via s.repo, use sql directly for now)
	type row struct{ CloudCoverMean float64 }
	db, ok := getDBFromRepo(s.repo)
	if !ok {
		return 0, 0, errors.New("repo does not expose *sql.DB")
	}
	query := `SELECT wd.cloud_cover_mean FROM actual_daily ad
		JOIN weather_daily wd ON ad.date = wd.date AND ad.solar_profile_id = wd.id
		WHERE ad.solar_profile_id = $1 AND ad.user_id = $2 AND ad.actual_kwh > 0 AND wd.cloud_cover_mean IS NOT NULL`
	rows, err := db.QueryContext(ctx, query, profileID, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("query site baseline: %w", err)
	}
	defer rows.Close()
	var sum float64
	var count int
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.CloudCoverMean); err != nil {
			continue
		}
		if r.CloudCoverMean >= 0 && r.CloudCoverMean <= 100 {
			sum += r.CloudCoverMean
			count++
		}
	}
	if count == 0 {
		return 0, 0, errors.New("no valid site baseline data")
	}
	avg := sum / float64(count)
	// 3. Cache result
	baseline := &WeatherBaseline{
		ProfileID:     profileID,
		UserID:        userID,
		BaselineType:  BaselineSite,
		BaselineValue: avg,
		SampleCount:   count,
		ValidFrom:     time.Now().AddDate(0, 0, -30),
		ValidTo:       time.Now().AddDate(0, 0, 1),
	}
	_ = s.repo.SaveBaseline(ctx, baseline)
	return avg, count, nil
}

// getDBFromRepo tries to extract *sql.DB from repo for direct query (hack for now)
func getDBFromRepo(repo interface{}) (*sql.DB, bool) {
	type dbGetter interface{ DB() *sql.DB }
	if getter, ok := repo.(dbGetter); ok {
		return getter.DB(), true
	}
	// fallback: try to access db field
	if r, ok := repo.(*struct{ DB *sql.DB }); ok {
		return r.DB, true
	}
	return nil, false
}

package weather

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Service defines business logic for fetching and caching weather data
type Service interface {
	FetchWeatherForDate(lat, lng float64, date time.Time) (*WeatherDaily, error)
}

type service struct {
	repo    Repository
	baseURL string
}

// NewService creates a new weather service
func NewService(repo Repository, baseURL string) Service {
	return &service{repo: repo, baseURL: baseURL}
}

// FetchWeatherForDate retrieves weather from cache or fetches it from Open-Meteo API
func (s *service) FetchWeatherForDate(lat, lng float64, date time.Time) (*WeatherDaily, error) {
	// Return cached data if already stored for this date and location
	cached, err := s.repo.GetWeatherByDateAndLocation(date, lat, lng)
	if err == nil {
		if cached != nil && cached.ShortwaveRadiationMJ > 0 {
			return cached, nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check weather cache: %w", err)
	}

	// Fetch fresh data from Open-Meteo
	w, err := s.fetchFromOpenMeteo(lat, lng, date)
	if err != nil {
		return nil, err
	}

	// Cache the result in the database
	if saveErr := s.repo.SaveWeatherDaily(w); saveErr != nil {
		fmt.Printf("warning: failed to cache weather data: %v\n", saveErr)
	}

	return w, nil
}

// fetchFromOpenMeteo calls the Open-Meteo API and maps the response to WeatherDaily
func (s *service) fetchFromOpenMeteo(lat, lng float64, date time.Time) (*WeatherDaily, error) {
	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf(
		"%s/forecast?latitude=%.4f&longitude=%.4f&daily=cloud_cover_mean,temperature_2m_mean,shortwave_radiation_sum&start_date=%s&end_date=%s&timezone=UTC",
		s.baseURL, lat, lng, dateStr, dateStr,
	)

	resp, err := http.Get(url) //nolint:gosec // URL is constructed from internal config, not user input
	if err != nil {
		return nil, fmt.Errorf("fetch weather from Open-Meteo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open-meteo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read weather response: %w", err)
	}

	var apiResp OpenMeteoResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parse weather response: %w", err)
	}

	if len(apiResp.Daily.CloudCoverMean) == 0 {
		return nil, fmt.Errorf("no weather data returned for %s", dateStr)
	}
	if len(apiResp.Daily.ShortwaveRadiationSum) == 0 {
		return nil, fmt.Errorf("no shortwave radiation data returned for %s", dateStr)
	}

	w := &WeatherDaily{
		ID:                   uuid.New(),
		Date:                 date,
		Lat:                  lat,
		Lng:                  lng,
		CloudCover:           int(apiResp.Daily.CloudCoverMean[0]),
		Temperature:          apiResp.Daily.Temperature2mMean[0],
		ShortwaveRadiationMJ: apiResp.Daily.ShortwaveRadiationSum[0],
		RawJSON:              string(body),
		CreatedAt:            time.Now().UTC(),
	}

	return w, nil
}

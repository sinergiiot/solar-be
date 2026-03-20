package weather

import (
	"time"

	"github.com/google/uuid"
)

// WeatherDaily holds a single day's weather data for a location
type WeatherDaily struct {
	ID                   uuid.UUID `json:"id"`
	Date                 time.Time `json:"date"`
	Lat                  float64   `json:"lat"`
	Lng                  float64   `json:"lng"`
	CloudCover           int       `json:"cloud_cover"`
	Temperature          float64   `json:"temperature"`
	ShortwaveRadiationMJ float64   `json:"shortwave_radiation_mj"`
	RawJSON              string    `json:"-"`
	CreatedAt            time.Time `json:"created_at"`
}

// OpenMeteoResponse maps the relevant fields from the Open-Meteo API response
type OpenMeteoResponse struct {
	Daily struct {
		Time                  []string  `json:"time"`
		CloudCoverMean        []float64 `json:"cloud_cover_mean"`
		Temperature2mMean     []float64 `json:"temperature_2m_mean"`
		ShortwaveRadiationSum []float64 `json:"shortwave_radiation_sum"`
	} `json:"daily"`
}

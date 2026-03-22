package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents an application user
type User struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	EmailVerified      bool      `json:"email_verified"`
	EmailVerifiedAt    *time.Time `json:"email_verified_at,omitempty"`
	PasswordHash       string    `json:"-"`
	ForecastEfficiency float64   `json:"forecast_efficiency"`
	CreatedAt          time.Time `json:"created_at"`
}

// CreateUserRequest holds data needed to register a new user
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service provides authentication workflows.
type Service interface {
	Register(name, email, password string) (*user.User, string, string, error)
	Login(email, password string) (*user.User, string, string, error)
	ParseToken(token string) (uuid.UUID, error)
	RefreshTokens(refreshToken string) (accessToken, newRefreshToken string, err error)
	RevokeRefreshToken(refreshToken string) error
}

type service struct {
	userService            user.Service
	refreshRepo            *refreshTokenRepository
	jwtSecret              []byte
	tokenExpiryHours       int
	refreshTokenExpiryDays int
}

// NewService creates auth service with dependencies.
func NewService(db *sql.DB, userService user.Service, jwtSecret string, tokenExpiryHours, refreshTokenExpiryDays int) Service {
	return &service{
		userService:            userService,
		refreshRepo:            newRefreshTokenRepository(db),
		jwtSecret:              []byte(jwtSecret),
		tokenExpiryHours:       tokenExpiryHours,
		refreshTokenExpiryDays: refreshTokenExpiryDays,
	}
}

// Register creates a user account and returns an access token + refresh token.
func (s *service) Register(name, email, password string) (*user.User, string, string, error) {
	u, err := s.userService.CreateUser(user.CreateUserRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.buildAccessToken(u.ID)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.issueRefreshToken(u.ID)
	if err != nil {
		return nil, "", "", err
	}

	return u, accessToken, refreshToken, nil
}

// Login validates credentials and returns an access token + refresh token.
func (s *service) Login(email, password string) (*user.User, string, string, error) {
	u, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", fmt.Errorf("invalid email or password")
		}
		return nil, "", "", err
	}

	if u.PasswordHash == "" {
		return nil, "", "", fmt.Errorf("account has no password configured")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", fmt.Errorf("invalid email or password")
	}

	accessToken, err := s.buildAccessToken(u.ID)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.issueRefreshToken(u.ID)
	if err != nil {
		return nil, "", "", err
	}

	return u, accessToken, refreshToken, nil
}

// ParseToken validates JWT and returns user id from subject.
func (s *service) ParseToken(token string) (uuid.UUID, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	subRaw, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("token subject is missing")
	}

	userID, err := uuid.Parse(subRaw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid token subject: %w", err)
	}

	return userID, nil
}

// RefreshTokens validates a refresh token, rotates it, and issues a new access token.
func (s *service) RefreshTokens(refreshToken string) (string, string, error) {
	row, err := s.refreshRepo.Find(refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("refresh token not found")
		}
		return "", "", fmt.Errorf("lookup refresh token: %w", err)
	}

	if row.revoked {
		return "", "", fmt.Errorf("refresh token has been revoked")
	}

	if time.Now().UTC().After(row.expiresAt) {
		return "", "", fmt.Errorf("refresh token has expired")
	}

	// Rotate: revoke the old token before issuing a new one.
	if err := s.refreshRepo.Revoke(refreshToken); err != nil {
		return "", "", fmt.Errorf("revoke old refresh token: %w", err)
	}

	newAccessToken, err := s.buildAccessToken(row.userID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.issueRefreshToken(row.userID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// RevokeRefreshToken marks a refresh token as revoked (used on logout).
func (s *service) RevokeRefreshToken(refreshToken string) error {
	return s.refreshRepo.Revoke(refreshToken)
}

// issueRefreshToken creates and stores a new refresh token for the given user.
func (s *service) issueRefreshToken(userID uuid.UUID) (string, error) {
	expiresAt := time.Now().UTC().Add(time.Duration(s.refreshTokenExpiryDays) * 24 * time.Hour)
	return s.refreshRepo.Create(userID, expiresAt)
}

// buildAccessToken issues a signed short-lived JWT for one user.
func (s *service) buildAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(time.Duration(s.tokenExpiryHours) * time.Hour)

	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": now.Unix(),
		"exp": exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

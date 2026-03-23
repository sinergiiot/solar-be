package auth

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// Service provides authentication workflows.
type Service interface {
	Register(name, email, password string) (*user.User, error)
	VerifyEmail(email, code string) (*user.User, string, string, error)
	ResendVerification(email string) error
	Login(email, password string) (*user.User, string, string, error)
	ParseToken(token string) (uuid.UUID, error)
	RefreshTokens(refreshToken string) (accessToken, newRefreshToken string, err error)
	RevokeRefreshToken(refreshToken string) error
	ForgotPassword(email string) error
	ResetPassword(email, code, newPassword string) error
}

type service struct {
	userService            user.Service
	refreshRepo            *refreshTokenRepository
	verificationRepo       *emailVerificationRepository
	passwordResetRepo      *passwordResetRepository
	jwtSecret              []byte
	tokenExpiryHours       int
	refreshTokenExpiryDays int
	verifyEmailOnRegister  bool
	smtpHost               string
	smtpPort               int
	smtpUsername           string
	smtpPassword           string
	smtpFrom               string
}

// NewService creates auth service with dependencies.
func NewService(
	db *sql.DB,
	userService user.Service,
	jwtSecret string,
	tokenExpiryHours,
	refreshTokenExpiryDays int,
	verifyEmailOnRegister bool,
	smtpHost,
	smtpPort,
	smtpUsername,
	smtpPassword,
	smtpFrom string,
) Service {
	parsedSMTPPort, err := strconv.Atoi(strings.TrimSpace(smtpPort))
	if err != nil || parsedSMTPPort <= 0 {
		parsedSMTPPort = 587
	}

	return &service{
		userService:            userService,
		refreshRepo:            newRefreshTokenRepository(db),
		verificationRepo:       newEmailVerificationRepository(db),
		passwordResetRepo:      newPasswordResetRepository(db),
		jwtSecret:              []byte(jwtSecret),
		tokenExpiryHours:       tokenExpiryHours,
		refreshTokenExpiryDays: refreshTokenExpiryDays,
		verifyEmailOnRegister:  verifyEmailOnRegister,
		smtpHost:               strings.TrimSpace(smtpHost),
		smtpPort:               parsedSMTPPort,
		smtpUsername:           strings.TrimSpace(smtpUsername),
		smtpPassword:           strings.TrimSpace(smtpPassword),
		smtpFrom:               strings.TrimSpace(smtpFrom),
	}
}

// Register creates a user account and sends one verification code to the registered email.
func (s *service) Register(name, email, password string) (*user.User, error) {
	if err := s.validateEmailAddress(email); err != nil {
		return nil, err
	}

	u, err := s.userService.CreateUser(user.CreateUserRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	if !s.verifyEmailOnRegister {
		if err := s.userService.MarkEmailVerified(u.ID); err != nil {
			return nil, fmt.Errorf("mark email verified: %w", err)
		}

		u.EmailVerified = true
		now := time.Now().UTC()
		u.EmailVerifiedAt = &now
		return u, nil
	}

	if err := s.sendAndStoreVerificationCode(u); err != nil {
		return nil, err
	}

	return u, nil
}

// VerifyEmail validates OTP code and returns access + refresh token for verified account.
func (s *service) VerifyEmail(email, code string) (*user.User, string, string, error) {
	u, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", fmt.Errorf("account not found")
		}
		return nil, "", "", err
	}

	if u.EmailVerified {
		accessToken, tokenErr := s.buildAccessToken(u.ID)
		if tokenErr != nil {
			return nil, "", "", tokenErr
		}

		refreshToken, tokenErr := s.issueRefreshToken(u.ID)
		if tokenErr != nil {
			return nil, "", "", tokenErr
		}

		return u, accessToken, refreshToken, nil
	}

	normalizedCode := strings.TrimSpace(code)
	if len(normalizedCode) != 6 {
		return nil, "", "", fmt.Errorf("verification code must be 6 digits")
	}

	row, err := s.verificationRepo.FindLatestByUserAndCode(u.ID, normalizedCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", fmt.Errorf("verification code is invalid")
		}
		return nil, "", "", fmt.Errorf("check verification code: %w", err)
	}

	if row.usedAt.Valid {
		return nil, "", "", fmt.Errorf("verification code has been used")
	}

	if time.Now().UTC().After(row.expiresAt) {
		return nil, "", "", fmt.Errorf("verification code has expired")
	}

	if err := s.verificationRepo.MarkCodeUsed(u.ID, normalizedCode); err != nil {
		return nil, "", "", fmt.Errorf("mark verification code used: %w", err)
	}

	if err := s.userService.MarkEmailVerified(u.ID); err != nil {
		return nil, "", "", fmt.Errorf("mark user email verified: %w", err)
	}

	u.EmailVerified = true
	now := time.Now().UTC()
	u.EmailVerifiedAt = &now

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

// ResendVerification generates and sends a fresh OTP code to one unverified account.
func (s *service) ResendVerification(email string) error {
	u, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("account not found")
		}
		return err
	}

	if u.EmailVerified {
		return nil
	}

	return s.sendAndStoreVerificationCode(u)
}

// validateEmailAddress validates address format and confirms MX records.
func (s *service) validateEmailAddress(rawEmail string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(rawEmail))
	parsedAddress, err := mail.ParseAddress(normalizedEmail)
	if err != nil || parsedAddress.Address == "" {
		return fmt.Errorf("email is not valid")
	}

	parts := strings.Split(parsedAddress.Address, "@")
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return fmt.Errorf("email domain is not valid")
	}

	domain := strings.TrimSpace(parts[1])
	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		return fmt.Errorf("email domain cannot receive emails")
	}

	return nil
}

// sendAndStoreVerificationCode issues one OTP and delivers it to user email.
func (s *service) sendAndStoreVerificationCode(u *user.User) error {

	if s.smtpHost == "" || s.smtpUsername == "" || s.smtpPassword == "" || s.smtpFrom == "" {
		return fmt.Errorf("email verification is unavailable: SMTP configuration is incomplete")
	}

	code, err := generateOTPCode()
	if err != nil {
		return fmt.Errorf("generate verification code: %w", err)
	}

	if err := s.verificationRepo.InvalidateAllByUser(u.ID); err != nil {
		return fmt.Errorf("invalidate previous verification codes: %w", err)
	}

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	if err := s.verificationRepo.Create(u.ID, code, expiresAt); err != nil {
		return fmt.Errorf("store verification code: %w", err)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.smtpFrom)
	message.SetHeader("To", u.Email)
	message.SetHeader("Subject", "Solar Forecast - Kode Verifikasi Email")
	message.SetBody("text/plain", fmt.Sprintf("Halo %s,\n\nKode verifikasi akun Solar Forecast Anda adalah: %s\nKode berlaku sampai %s (UTC).\n\nJika Anda tidak melakukan registrasi, abaikan email ini.", u.Name, code, expiresAt.Format("15:04:05 2006-01-02")))

	dialer := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)
	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to deliver verification email: %w", err)
	}

	return nil
}

// generateOTPCode creates one random 6-digit numeric code.
func generateOTPCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
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

	if !u.EmailVerified {
		return nil, "", "", fmt.Errorf("email is not verified")
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

// ForgotPassword generates a reset code and sends it via email.
func (s *service) ForgotPassword(email string) error {
	u, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("account with this email not found")
		}
		return err
	}

	if s.smtpHost == "" || s.smtpUsername == "" || s.smtpPassword == "" || s.smtpFrom == "" {
		return fmt.Errorf("password reset is unavailable: SMTP configuration is incomplete")
	}

	code, err := generateOTPCode()
	if err != nil {
		return fmt.Errorf("generate reset code: %w", err)
	}

	if err := s.passwordResetRepo.InvalidateAllByUser(u.ID); err != nil {
		return fmt.Errorf("invalidate previous reset codes: %w", err)
	}

	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	if err := s.passwordResetRepo.Create(u.ID, code, expiresAt); err != nil {
		return fmt.Errorf("store reset code: %w", err)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.smtpFrom)
	message.SetHeader("To", u.Email)
	message.SetHeader("Subject", "Solar Forecast - Reset Password")
	message.SetBody("text/plain", fmt.Sprintf("Halo %s,\n\nKami menerima permintaan untuk reset password akun Solar Forecast Anda.\nKode reset Anda adalah: %s\nKode berlaku sampai %s (UTC).\n\nJika Anda tidak merasa melakukan permintaan ini, abaikan email ini.", u.Name, code, expiresAt.Format("15:04:05 2006-01-02")))

	dialer := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)
	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to deliver reset email: %w", err)
	}

	return nil
}

// ResetPassword validates OTP and updates user password.
func (s *service) ResetPassword(email, code, newPassword string) error {
	u, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("account not found")
		}
		return err
	}

	normalizedCode := strings.TrimSpace(code)
	if len(normalizedCode) != 6 {
		return fmt.Errorf("verification code must be 6 digits")
	}

	row, err := s.passwordResetRepo.FindLatestByUserAndCode(u.ID, normalizedCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("reset code is invalid")
		}
		return fmt.Errorf("check reset code: %w", err)
	}

	if row.usedAt.Valid {
		return fmt.Errorf("reset code has been used")
	}

	if time.Now().UTC().After(row.expiresAt) {
		return fmt.Errorf("reset code has expired")
	}

	if err := s.userService.UpdatePassword(u.ID, newPassword); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := s.passwordResetRepo.MarkCodeUsed(u.ID, normalizedCode); err != nil {
		log.Printf("failed to mark reset code as used for user %s: %v", u.ID, err)
	}

	return nil
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

package auth

import (
	"encoding/json"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler exposes auth HTTP endpoints.
type Handler struct {
	authService Service
	userService user.Service
}

// NewHandler creates auth handler.
func NewHandler(authService Service, userService user.Service) *Handler {
	return &Handler{authService: authService, userService: userService}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// RegisterPublicRoutes wires public auth endpoints.
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/verify-email", h.VerifyEmail)
	r.Post("/auth/resend-verification", h.ResendVerification)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/forgot-password", h.ForgotPassword)
	r.Post("/auth/reset-password", h.ResetPassword)
}

// RegisterProtectedRoutes wires authenticated auth endpoints.
func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Get("/auth/me", h.Me)
}

// Register handles POST /auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	u, err := h.authService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "Akun berhasil dibuat. Silakan cek email untuk kode verifikasi.",
		"verification_required": true,
		"user": map[string]any{
			"id":             u.ID,
			"name":           u.Name,
			"email":          u.Email,
			"email_verified": u.EmailVerified,
		},
	})
}

// VerifyEmail handles POST /auth/verify-email.
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	u, accessToken, refreshToken, err := h.authService.VerifyEmail(req.Email, req.Code)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"id":             u.ID,
			"name":           u.Name,
			"email":          u.Email,
			"email_verified": u.EmailVerified,
		},
	})
}

// ResendVerification handles POST /auth/resend-verification.
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.authService.ResendVerification(req.Email); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Kode verifikasi baru telah dikirim."})
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	u, accessToken, refreshToken, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"id":             u.ID,
			"name":           u.Name,
			"email":          u.Email,
			"email_verified": u.EmailVerified,
		},
	})
}

// Logout handles POST /auth/logout — revokes the provided refresh token.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.RefreshToken != "" {
		// Best-effort revocation; ignore errors so logout always succeeds client-side.
		_ = h.authService.RevokeRefreshToken(body.RefreshToken)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Refresh handles POST /auth/refresh — issues new tokens using a valid refresh token.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
		return
	}

	accessToken, newRefreshToken, err := h.authService.RefreshTokens(body.RefreshToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

// Me handles GET /auth/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	u, err := h.userService.GetUserByID(userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":                  u.ID,
		"name":                u.Name,
		"email":               u.Email,
		"email_verified":      u.EmailVerified,
		"forecast_efficiency": u.ForecastEfficiency,
	})
}

// ForgotPassword handles POST /auth/forgot-password.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.authService.ForgotPassword(req.Email); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Kode reset password telah dikirim ke email Anda."})
}

// ResetPassword handles POST /auth/reset-password.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.authService.ResetPassword(req.Email, req.Code, req.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password Anda berhasil diperbarui. Silakan login kembali."})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

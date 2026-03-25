package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/akbarsenawijaya/solar-forecast/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the user domain
type Handler struct {
	service Service
}

// NewHandler creates a new user HTTP handler
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes wires all user endpoints onto the given router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/users", h.CreateUser)
	r.Get("/users", h.GetAllUsers)
	r.Get("/users/{id}", h.GetUserByID)

	// Expects this to be mounted on a protected subrouter
	r.Post("/users/me/branding", h.UpdateBranding)
	// E5-T6: ESG Share
	r.Get("/users/me/esg-share", h.GetESGShareStatus)
	r.Post("/users/me/esg-share/enable", h.EnableESGShare)
	r.Post("/users/me/esg-share/disable", h.DisableESGShare)
}

// CreateUser handles POST /users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.service.CreateUser(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

// GetAllUsers handles GET /users
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// GetUserByID handles GET /users/{id}
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	u, err := h.service.GetUserByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, u)
}
// UpdateBranding handles POST /users/me/branding
func (h *Handler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	// Require userID from protected context
	userID, ok := r.Context().Value(ctxkeys.UserID).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	companyName := r.FormValue("company_name")
	
	file, header, err := r.FormFile("logo")
	logoURL := ""
	if err == nil {
		defer file.Close()
		
		// Create uploads dir
		if err := os.MkdirAll("uploads/logos", 0755); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create directory")
			return
		}
		
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".png" // fallback
		}
		
		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		path := filepath.Join("uploads", "logos", filename)
		
		outFile, err := os.Create(path)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save logo")
			return
		}
		defer outFile.Close()
		
		if _, err := io.Copy(outFile, file); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to write logo")
			return
		}
		
		logoURL = "/uploads/logos/" + filename
	} else if err != http.ErrMissingFile {
		writeError(w, http.StatusBadRequest, "failed to read logo file")
		return
	} else {
		// If no new logo is provided, we should probably not overwrite with empty?
		// Check current user logic
		user, getErr := h.service.GetUserByID(userID)
		if getErr == nil {
			logoURL = user.CompanyLogoURL
		}
	}

	if err := h.service.UpdateBranding(userID, companyName, logoURL); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"company_name": companyName,
		"company_logo_url": logoURL,
	})
}

// writeJSON encodes v as JSON and writes it to the response
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError sends a JSON error response
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GetESGShareStatus handles GET /users/me/esg-share
func (h *Handler) GetESGShareStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkeys.UserID).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	u, err := h.service.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": u.ESGShareEnabled,
		"token":   u.ESGShareToken,
	})
}

// EnableESGShare handles POST /users/me/esg-share/enable
func (h *Handler) EnableESGShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkeys.UserID).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	token, err := h.service.EnableESGShare(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": true,
		"token":   token,
	})
}

// DisableESGShare handles POST /users/me/esg-share/disable
func (h *Handler) DisableESGShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkeys.UserID).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.service.DisableESGShare(userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
}

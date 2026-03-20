package user

import (
	"encoding/json"
	"net/http"

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

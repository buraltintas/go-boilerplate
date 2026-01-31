package auth

import (
	"encoding/json"
	"net/http"

	"go-boilerplate/internal/contextutil"
	"go-boilerplate/internal/errors"
	"go-boilerplate/internal/response"
)

// Handler handles HTTP requests for authentication.
// Single responsibility: HTTP transport layer for auth operations.
// Design: Handlers translate HTTP → domain, call service, translate domain → HTTP.
// NO business logic here.
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register handles user registration.
// POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errors.BadRequest("invalid request body"))
		return
	}

	authResp, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.JSON(w, http.StatusCreated, authResp)
}

// Login handles user authentication.
// POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errors.BadRequest("invalid request body"))
		return
	}

	authResp, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, authResp)
}

// RefreshToken handles token refresh.
// POST /auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errors.BadRequest("invalid request body"))
		return
	}

	authResp, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, authResp)
}

// Logout handles user logout.
// POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errors.BadRequest("invalid request body"))
		return
	}

	if err := h.service.Logout(req.RefreshToken); err != nil {
		response.Error(w, r, err)
		return
	}

	response.NoContent(w)
}

// Me returns the current authenticated user.
// GET /auth/me
// Requires authentication middleware.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := contextutil.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	user, err := h.service.GetCurrentUser(userID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, user)
}

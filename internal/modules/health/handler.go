package health

import (
	"net/http"

	"go-boilerplate/internal/response"
)

// Handler handles health check requests.
// Single responsibility: health check endpoint.
type Handler struct{}

// NewHandler creates a new health handler.
func NewHandler() *Handler {
	return &Handler{}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// Check handles health check requests.
// GET /health
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, HealthResponse{
		Status: "healthy",
	})
}

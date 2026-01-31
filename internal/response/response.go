package response

import (
	"encoding/json"
	"net/http"

	"go-boilerplate/internal/errors"
	"go-boilerplate/internal/observability"
)

// Response envelope for consistent API responses.
// Single responsibility: standardize HTTP response format.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON sends a successful JSON response.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Response{
		Success: true,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		observability.Error("failed to encode response", "error", err)
	}
}

// Error sends an error JSON response.
// Accepts AppError for typed errors, or generic errors.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// Type assertion to check if it's our AppError
	appErr, ok := err.(*errors.AppError)
	if !ok {
		// Unknown error - don't leak internal details
		observability.Error("unhandled error", "error", err, "request_id", r.Context().Value("request_id"))
		appErr = errors.ErrInternalServer
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)

	resp := Response{
		Success: false,
		Error: &ErrorData{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		observability.Error("failed to encode error response", "error", err)
	}
}

// NoContent sends a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

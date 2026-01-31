package middleware

import (
	"context"
	"net/http"

	"go-boilerplate/internal/security"
)

// RequestID generates a unique request ID for each request.
// Single responsibility: request correlation.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a secure random request ID
		requestID, err := security.GenerateSecureToken(16)
		if err != nil {
			// Fallback to a simple ID if generation fails
			requestID = "unknown"
		}

		// Add to response header for client tracking
		w.Header().Set("X-Request-ID", requestID)

		// Add to context for use in handlers and logging
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

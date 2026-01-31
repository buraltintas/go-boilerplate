package middleware

import (
	"net/http"

	"go-boilerplate/internal/errors"
	"go-boilerplate/internal/observability"
	"go-boilerplate/internal/response"
)

// Recovery recovers from panics and returns a 500 error.
// Single responsibility: panic recovery.
// This must be the outermost middleware in the chain.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with context
				requestID := r.Context().Value("request_id")
				observability.Error(
					"panic recovered",
					"error", err,
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
				)

				// Return generic error to client (don't leak panic details)
				response.Error(w, r, errors.ErrInternalServer)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

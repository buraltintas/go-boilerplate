package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-boilerplate/internal/errors"
	"go-boilerplate/internal/observability"
	"go-boilerplate/internal/response"
	"go-boilerplate/internal/security"
)

// TokenValidator defines the interface for token validation.
// Interface Segregation Principle: small, focused interface.
type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*security.Claims, error)
}

// Auth validates JWT tokens and injects user identity into context.
// Single responsibility: authentication middleware.
// Design: Permission checks belong in services, not here.
func Auth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, r, errors.ErrMissingToken)
				return
			}

			// Check Bearer scheme
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, r, errors.BadRequest("invalid authorization header format"))
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := validator.ValidateAccessToken(tokenString)
			if err != nil {
				observability.Warn(
					"invalid token",
					"error", err,
					"request_id", r.Context().Value("request_id"),
				)
				response.Error(w, r, errors.ErrInvalidToken)
				return
			}

			// Inject user identity into context
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

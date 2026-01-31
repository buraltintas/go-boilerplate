package contextutil

import (
	"context"

	"go-boilerplate/internal/errors"
)

// GetUserID extracts the authenticated user ID from context.
// Single responsibility: context value extraction with type safety.
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", errors.Unauthorized("user not authenticated")
	}
	return userID, nil
}

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value("request_id").(string)
	if !ok {
		return "unknown"
	}
	return requestID
}

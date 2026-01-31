package errors

import "net/http"

// AppError represents a typed application error.
// Single responsibility: standardize error handling across the application.
type AppError struct {
	Code    string // machine-readable error code
	Message string // human-readable message (safe to expose to clients)
	Status  int    // HTTP status code
}

func (e *AppError) Error() string {
	return e.Message
}

// Predefined error constructors for common scenarios

func BadRequest(message string) *AppError {
	return &AppError{
		Code:    "BAD_REQUEST",
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Code:    "FORBIDDEN",
		Message: message,
		Status:  http.StatusForbidden,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: message,
		Status:  http.StatusNotFound,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Status:  http.StatusConflict,
	}
}

func Internal(message string) *AppError {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}

func TooManyRequests(message string) *AppError {
	return &AppError{
		Code:    "TOO_MANY_REQUESTS",
		Message: message,
		Status:  http.StatusTooManyRequests,
	}
}

// Common predefined errors
var (
	ErrInvalidCredentials = Unauthorized("invalid credentials")
	ErrInvalidToken       = Unauthorized("invalid or expired token")
	ErrMissingToken       = Unauthorized("authorization token required")
	ErrInternalServer     = Internal("internal server error")
)

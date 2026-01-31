package middleware

import (
	"net/http"
	"sync"
	"time"

	"go-boilerplate/internal/errors"
	"go-boilerplate/internal/observability"
	"go-boilerplate/internal/response"
)

// RateLimiter implements simple in-memory rate limiting.
// Single responsibility: prevent abuse via request rate limiting.
//
// Design notes:
// - This is intentionally simple and readable
// - Uses IP-based limiting
// - In-memory storage (will reset on restart)
// - NOT suitable for horizontal scaling
// - Can be replaced with Redis-backed limiter later without changing interface
type RateLimiter struct {
	requestsPerMinute int
	enabled           bool
	clients           map[string]*clientLimiter
	mu                sync.RWMutex
	cleanupInterval   time.Duration
}

// clientLimiter tracks requests for a single client.
type clientLimiter struct {
	requests  int
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(requestsPerMinute int, enabled bool) *RateLimiter {
	rl := &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		enabled:           enabled,
		clients:           make(map[string]*clientLimiter),
		cleanupInterval:   5 * time.Minute,
	}

	// Start background cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns the rate limiting middleware.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract client identifier (IP address)
		clientIP := getClientIP(r)

		// Check rate limit
		if !rl.allowRequest(clientIP) {
			observability.Warn(
				"rate limit exceeded",
				"client_ip", clientIP,
				"path", r.URL.Path,
				"request_id", r.Context().Value("request_id"),
			)
			response.Error(w, r, errors.TooManyRequests("rate limit exceeded"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allowRequest checks if a request should be allowed.
func (rl *RateLimiter) allowRequest(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	limiter, exists := rl.clients[clientIP]
	if !exists || now.After(limiter.resetTime) {
		// Create new limiter for this client
		rl.clients[clientIP] = &clientLimiter{
			requests:  1,
			resetTime: now.Add(time.Minute),
		}
		return true
	}

	// Check if limit exceeded
	if limiter.requests >= rl.requestsPerMinute {
		return false
	}

	// Increment request count
	limiter.requests++
	return true
}

// cleanup removes expired client entries periodically.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, limiter := range rl.clients {
			if now.After(limiter.resetTime) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client's IP address from the request.
// Checks X-Forwarded-For and X-Real-IP headers first (for proxies/load balancers).
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := len(xff); idx > 0 {
			if idx := len(xff); idx > 0 {
				return xff[:idx]
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

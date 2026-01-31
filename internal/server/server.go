package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/middleware"
	"go-boilerplate/internal/modules/auth"
	"go-boilerplate/internal/modules/health"
	"go-boilerplate/internal/observability"
	"go-boilerplate/internal/realtime"
	"go-boilerplate/internal/security"
)

// Server represents the HTTP server with all dependencies.
// Single responsibility: server lifecycle management and dependency wiring.
type Server struct {
	config       *config.Config
	httpServer   *http.Server
	hub          *realtime.Hub
	tokenManager *security.TokenManager
}

// New creates a new server instance.
// This is where dependency injection happens - everything is explicit.
func New(cfg *config.Config) *Server {
	// Initialize dependencies
	tokenManager := security.NewTokenManager(
		cfg.Auth.AccessTokenSecret,
		cfg.Auth.RefreshTokenSecret,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
	)

	hub := realtime.NewHub()

	// Create server
	srv := &Server{
		config:       cfg,
		hub:          hub,
		tokenManager: tokenManager,
	}

	// Setup router
	router := srv.setupRouter()

	// Create HTTP server with timeouts
	srv.httpServer = &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return srv
}

// setupRouter configures all routes and middleware.
// Middleware order is explicit and intentional.
func (s *Server) setupRouter() http.Handler {
	mux := http.NewServeMux()

	// Initialize handlers
	healthHandler := health.NewHandler()
	authHandler := s.createAuthHandler()
	wsHandler := realtime.NewHandler(s.hub)

	// Public routes
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("POST /v1/auth/logout", authHandler.Logout)

	// Protected routes
	authMiddleware := middleware.Auth(s.tokenManager)
	mux.Handle("GET /v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /v1/ws", authMiddleware(http.HandlerFunc(wsHandler.Connect)))

	// Apply global middleware chain
	// Order matters: outermost first
	handler := middleware.Recovery(mux)                         // 1. Catch panics
	handler = middleware.RequestID(handler)                     // 2. Add request ID
	handler = middleware.Logger(handler)                        // 3. Log requests
	handler = middleware.CORS(middleware.CORSConfig{            // 4. CORS
		AllowedOrigins: s.config.CORS.AllowedOrigins,
		AllowedMethods: s.config.CORS.AllowedMethods,
		AllowedHeaders: s.config.CORS.AllowedHeaders,
	})(handler)
	handler = middleware.NewRateLimiter(                        // 5. Rate limiting
		s.config.Rate.RequestsPerMinute,
		s.config.Rate.Enabled,
	).Middleware(handler)

	return handler
}

// createAuthHandler creates the auth handler with all dependencies.
func (s *Server) createAuthHandler() *auth.Handler {
	// In-memory user store (will be replaced with database later)
	userStore := auth.NewInMemoryUserStore()

	// Auth service
	authService := auth.NewService(s.tokenManager, userStore)

	// Auth handler
	return auth.NewHandler(authService)
}

// Start starts the server and blocks until shutdown.
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.hub.Run()

	// Start HTTP server in a goroutine
	go func() {
		observability.Info(
			"server starting",
			"port", s.config.Server.Port,
			"read_timeout", s.config.Server.ReadTimeout,
			"write_timeout", s.config.Server.WriteTimeout,
		)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			observability.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	s.waitForShutdown()

	return nil
}

// waitForShutdown blocks until an interrupt signal is received, then gracefully shuts down.
func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	observability.Info("server shutting down")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(ctx); err != nil {
		observability.Error("server forced to shutdown", "error", err)
	}

	observability.Info("server stopped")
}

// GetHub returns the WebSocket hub (useful for triggering events from services).
func (s *Server) GetHub() *realtime.Hub {
	return s.hub
}

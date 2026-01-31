package auth

import (
	"fmt"
	"sync"

	"go-boilerplate/internal/errors"
	"go-boilerplate/internal/security"
)

// Service handles authentication business logic.
// Single responsibility: authentication operations.
// Design: This is where business rules live, not in handlers.
type Service struct {
	tokenManager   *security.TokenManager
	userStore      UserStore
	refreshTokens  map[string]string // token -> userID (in-memory for now)
	refreshTokenMu sync.RWMutex
}

// UserStore defines the interface for user persistence.
// Dependency Inversion Principle: depend on abstractions.
// Interface Segregation Principle: minimal, focused interface.
type UserStore interface {
	FindByEmail(email string) (*User, error)
	Create(user *User) error
	FindByID(id string) (*User, error)
}

// NewService creates a new auth service.
func NewService(tokenManager *security.TokenManager, userStore UserStore) *Service {
	return &Service{
		tokenManager:  tokenManager,
		userStore:     userStore,
		refreshTokens: make(map[string]string),
	}
}

// Register creates a new user account.
func (s *Service) Register(email, password string) (*AuthResponse, error) {
	// Validate input
	if email == "" || password == "" {
		return nil, errors.BadRequest("email and password are required")
	}

	// Check if user already exists
	existing, _ := s.userStore.FindByEmail(email)
	if existing != nil {
		return nil, errors.Conflict("user already exists")
	}

	// Hash password
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID
	userID, err := security.GenerateSecureToken(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Create user
	user := &User{
		ID:           userID,
		Email:        email,
		PasswordHash: hashedPassword,
	}

	if err := s.userStore.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.tokenManager.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token
	s.storeRefreshToken(tokenPair.RefreshToken, user.ID)

	return &AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Login authenticates a user and returns tokens.
func (s *Service) Login(email, password string) (*AuthResponse, error) {
	// Validate input
	if email == "" || password == "" {
		return nil, errors.BadRequest("email and password are required")
	}

	// Find user
	user, err := s.userStore.FindByEmail(email)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Verify password
	if err := security.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Generate tokens
	tokenPair, err := s.tokenManager.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token
	s.storeRefreshToken(tokenPair.RefreshToken, user.ID)

	return &AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken generates a new access token from a refresh token.
func (s *Service) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// Check if refresh token is stored (not revoked)
	if !s.isRefreshTokenValid(refreshToken, claims.UserID) {
		return nil, errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userStore.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.NotFound("user not found")
	}

	// Generate new token pair
	tokenPair, err := s.tokenManager.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Remove old refresh token and store new one
	s.revokeRefreshToken(refreshToken)
	s.storeRefreshToken(tokenPair.RefreshToken, user.ID)

	return &AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Logout invalidates a refresh token.
func (s *Service) Logout(refreshToken string) error {
	s.revokeRefreshToken(refreshToken)
	return nil
}

// GetCurrentUser retrieves the authenticated user.
func (s *Service) GetCurrentUser(userID string) (*UserResponse, error) {
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		return nil, errors.NotFound("user not found")
	}
	return user.ToResponse(), nil
}

// Refresh token management (in-memory for now)

func (s *Service) storeRefreshToken(token, userID string) {
	s.refreshTokenMu.Lock()
	defer s.refreshTokenMu.Unlock()
	s.refreshTokens[token] = userID
}

func (s *Service) isRefreshTokenValid(token, userID string) bool {
	s.refreshTokenMu.RLock()
	defer s.refreshTokenMu.RUnlock()
	storedUserID, exists := s.refreshTokens[token]
	return exists && storedUserID == userID
}

func (s *Service) revokeRefreshToken(token string) {
	s.refreshTokenMu.Lock()
	defer s.refreshTokenMu.Unlock()
	delete(s.refreshTokens, token)
}

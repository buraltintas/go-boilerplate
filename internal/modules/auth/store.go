package auth

import (
	"fmt"
	"sync"
)

// InMemoryUserStore is a simple in-memory user store for demonstration.
// Single responsibility: user persistence (in-memory).
// Design: This will be replaced with a real database implementation later.
type InMemoryUserStore struct {
	users   map[string]*User // userID -> User
	emails  map[string]string // email -> userID
	mu      sync.RWMutex
}

// NewInMemoryUserStore creates a new in-memory user store.
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users:  make(map[string]*User),
		emails: make(map[string]string),
	}
}

// FindByEmail finds a user by email.
func (s *InMemoryUserStore) FindByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, exists := s.emails[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	user, exists := s.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// Create creates a new user.
func (s *InMemoryUserStore) Create(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if email already exists
	if _, exists := s.emails[user.Email]; exists {
		return fmt.Errorf("user already exists")
	}

	s.users[user.ID] = user
	s.emails[user.Email] = user.ID
	return nil
}

// FindByID finds a user by ID.
func (s *InMemoryUserStore) FindByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

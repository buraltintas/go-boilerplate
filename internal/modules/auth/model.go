package auth

// User represents a user in the system.
// Single responsibility: user data representation.
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // never expose password hash in JSON
}

// LoginRequest represents login credentials.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents registration data.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents authentication response.
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
}

// UserResponse represents user data in responses (excludes sensitive fields).
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// ToResponse converts User to UserResponse.
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:    u.ID,
		Email: u.Email,
	}
}

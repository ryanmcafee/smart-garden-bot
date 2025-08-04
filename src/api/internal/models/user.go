package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email" validate:"required,email"`
	Name         string     `json:"name" db:"name" validate:"required,min=2,max=255"`
	Auth0UserID  *string    `json:"auth0_user_id,omitempty" db:"auth0_user_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Metadata     Metadata   `json:"metadata,omitempty" db:"metadata"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email       string   `json:"email" validate:"required,email"`
	Name        string   `json:"name" validate:"required,min=2,max=255"`
	Auth0UserID *string  `json:"auth0_user_id,omitempty"`
	Metadata    Metadata `json:"metadata,omitempty"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Name     *string  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Metadata Metadata `json:"metadata,omitempty"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Metadata  Metadata  `json:"metadata,omitempty"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Metadata:  u.Metadata,
	}
}

// APIKey represents an API key for machine-to-machine authentication
type APIKey struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name" validate:"required,min=3,max=255"`
	KeyHash     string    `json:"-" db:"key_hash"` // Never expose in JSON
	KeyPrefix   string    `json:"key_prefix" db:"key_prefix"`
	Scopes      []string  `json:"scopes" db:"scopes"`
	Permissions []string  `json:"permissions" db:"permissions"` // For backwards compatibility
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" validate:"required,min=3,max=255"`
	Scopes      []string   `json:"scopes,omitempty"`
	Permissions []string   `json:"permissions,omitempty"` // For backwards compatibility
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"` // Only returned once during creation
	KeyPrefix   string    `json:"key_prefix"`
	Permissions []string  `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserSession represents an active user session
type UserSession struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	TokenID          string     `json:"token_id" db:"token_id"`
	RefreshTokenHash *string    `json:"-" db:"refresh_token_hash"` // Never expose
	IPAddress        *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent        *string    `json:"user_agent,omitempty" db:"user_agent"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	LastActivityAt   time.Time  `json:"last_activity_at" db:"last_activity_at"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"` // seconds
	TokenType    string        `json:"token_type"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	SessionID string    `json:"session_id"`
	Scopes    []string  `json:"scopes,omitempty"`
}

// Session represents a user session
type Session struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"` // Never expose
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`
	NewUsers    int `json:"new_users"`
}
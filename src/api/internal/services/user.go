package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/database"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user-related operations
type UserService struct {
	db *database.DB
}

// NewUserService creates a new user service
func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, name, auth0_user_id, created_at, updated_at, deleted_at, metadata
		FROM auth.users
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	var user models.User
	err := s.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, auth0_user_id, created_at, updated_at, deleted_at, metadata
		FROM auth.users
		WHERE email = $1 AND deleted_at IS NULL
	`
	
	var user models.User
	err := s.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return &user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		ID:          uuid.New(),
		Email:       req.Email,
		Name:        req.Name,
		Auth0UserID: req.Auth0UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}
	
	query := `
		INSERT INTO auth.users (id, email, name, auth0_user_id, created_at, updated_at, metadata)
		VALUES (:id, :email, :name, :auth0_user_id, :created_at, :updated_at, :metadata)
	`
	
	_, err := s.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	// Build dynamic query based on provided fields
	setParts := []string{}
	args := map[string]interface{}{
		"user_id":    userID,
		"updated_at": time.Now(),
	}
	
	if req.Name != nil {
		setParts = append(setParts, "name = :name")
		args["name"] = *req.Name
	}
	
	if req.Metadata != nil {
		setParts = append(setParts, "metadata = :metadata")
		args["metadata"] = req.Metadata
	}
	
	if len(setParts) == 0 {
		return s.GetUserByID(ctx, userID) // No changes, return existing user
	}
	
	// Always update updated_at
	setParts = append(setParts, "updated_at = :updated_at")
	
	query := fmt.Sprintf(`
		UPDATE auth.users 
		SET %s
		WHERE id = :user_id AND deleted_at IS NULL
	`, strings.Join(setParts, ", "))
	
	_, err := s.db.NamedExecContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return s.GetUserByID(ctx, userID)
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth.users 
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	
	result, err := s.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// CreateAPIKey creates a new API key for a user
func (s *UserService) CreateAPIKey(ctx context.Context, userID uuid.UUID, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	// Generate a random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	
	key := hex.EncodeToString(keyBytes)
	keyPrefix := key[:8]
	
	// Hash the key for storage
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])
	
	apiKey := models.APIKey{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		KeyPrefix:   keyPrefix,
		KeyHash:     keyHash,
		Permissions: req.Permissions,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	query := `
		INSERT INTO auth.api_keys (id, user_id, name, key_prefix, key_hash, permissions, expires_at, created_at, updated_at)
		VALUES (:id, :user_id, :name, :key_prefix, :key_hash, :permissions, :expires_at, :created_at, :updated_at)
	`
	
	_, err := s.db.NamedExecContext(ctx, query, &apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}
	
	return &models.CreateAPIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Key:         key, // Return the full key only on creation
		KeyPrefix:   keyPrefix,
		Permissions: apiKey.Permissions,
		ExpiresAt:   apiKey.ExpiresAt,
		CreatedAt:   apiKey.CreatedAt,
	}, nil
}

// ListAPIKeys retrieves all API keys for a user
func (s *UserService) ListAPIKeys(ctx context.Context, userID uuid.UUID) ([]models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, permissions, expires_at, last_used_at, created_at, updated_at
		FROM auth.api_keys
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`
	
	var apiKeys []models.APIKey
	err := s.db.SelectContext(ctx, &apiKeys, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	
	return apiKeys, nil
}

// RevokeAPIKey revokes an API key
func (s *UserService) RevokeAPIKey(ctx context.Context, userID, apiKeyID uuid.UUID) error {
	query := `
		UPDATE auth.api_keys 
		SET revoked_at = $1, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND revoked_at IS NULL
	`
	
	result, err := s.db.ExecContext(ctx, query, time.Now(), apiKeyID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	
	return nil
}

// ValidatePassword validates a user's password
func (s *UserService) ValidatePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// HashPassword hashes a password using bcrypt
func (s *UserService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CreateSession creates a new user session
func (s *UserService) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string) (*models.Session, error) {
	session := &models.Session{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	query := `
		INSERT INTO auth.sessions (id, user_id, token_hash, expires_at, created_at, updated_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :created_at, :updated_at)
	`
	
	_, err := s.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	return session, nil
}

// GetSessionByTokenHash retrieves a session by token hash
func (s *UserService) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, updated_at
		FROM auth.sessions
		WHERE token_hash = $1 AND expires_at > NOW() AND revoked_at IS NULL
	`
	
	var session models.Session
	err := s.db.GetContext(ctx, &session, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return &session, nil
}

// RevokeSession revokes a session
func (s *UserService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE auth.sessions 
		SET revoked_at = $1, updated_at = $1
		WHERE id = $2 AND revoked_at IS NULL
	`
	
	_, err := s.db.ExecContext(ctx, query, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	
	return nil
}

// CreateUserSession creates a new user session with extended parameters
func (s *UserService) CreateUserSession(ctx context.Context, userID uuid.UUID, tokenID string, refreshToken *string, ipAddress string, userAgent string, expiresAt time.Time) (*models.Session, error) {
	session := &models.Session{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenID, // In a real implementation, this should be hashed
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	query := `
		INSERT INTO auth.sessions (id, user_id, token_hash, expires_at, created_at, updated_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :created_at, :updated_at)
	`
	
	_, err := s.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}
	
	return session, nil
}

// RevokeUserSession revokes a user session (alias for RevokeSession)
func (s *UserService) RevokeUserSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.RevokeSession(ctx, sessionID)
}

// GetUserStats returns user statistics (admin only)
func (s *UserService) GetUserStats(ctx context.Context) (*models.UserStats, error) {
	stats := &models.UserStats{}
	
	// Get total users
	err := s.db.GetContext(ctx, &stats.TotalUsers,
		"SELECT COUNT(*) FROM auth.users WHERE deleted_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}
	
	// Get active users (users with sessions in the last 30 days)
	err = s.db.GetContext(ctx, &stats.ActiveUsers,
		`SELECT COUNT(DISTINCT user_id) FROM auth.sessions 
		 WHERE created_at >= NOW() - INTERVAL '30 days'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	
	// Get new users (created in the last 30 days)
	err = s.db.GetContext(ctx, &stats.NewUsers,
		`SELECT COUNT(*) FROM auth.users 
		 WHERE created_at >= NOW() - INTERVAL '30 days' AND deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to get new users: %w", err)
	}
	
	return stats, nil
}
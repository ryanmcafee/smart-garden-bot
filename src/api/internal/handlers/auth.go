package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/config"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/middleware"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userService *services.UserService
	jwtConfig   config.JWTConfig
	validator   *validator.Validate
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *services.UserService, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtConfig:   jwtConfig,
		validator:   validator.New(),
	}
}

// Login handles user login requests
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: getValidationErrors(err),
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid email or password",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Note: In production, you would validate password against stored hash
	// For this example, we'll assume Auth0 handles authentication
	// and this endpoint is mainly for local/development authentication

	// Generate JWT token
	tokenString, err := middleware.GenerateToken(
		user.ID.String(),
		user.Email,
		models.RoleUser,
		h.jwtConfig,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Failed to generate access token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Create refresh token (simplified - in production use proper refresh token generation)
	refreshToken, err := middleware.GenerateToken(
		user.ID.String(),
		user.Email,
		models.RoleUser,
		config.JWTConfig{
			Secret:     h.jwtConfig.Secret,
			Expiration: time.Hour * 24 * 7, // 7 days
			Issuer:     h.jwtConfig.Issuer,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Failed to generate refresh token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Create user session
	tokenID := uuid.New().String()
	_, err = h.userService.CreateUserSession(
		c.Request.Context(),
		user.ID,
		tokenID,
		&refreshToken,
		c.ClientIP(),
		c.Request.UserAgent(),
		time.Now().Add(h.jwtConfig.Expiration),
	)
	if err != nil {
		// Log error but don't fail the login
		// In production, consider if this should be a hard failure
	}

	response := models.LoginResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(h.jwtConfig.Expiration.Seconds()),
		TokenType:    "Bearer",
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		Timestamp: time.Now(),
	})
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Validate refresh token
	claims, err := middleware.ValidateToken(req.RefreshToken, h.jwtConfig)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: "Invalid or expired refresh token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_TOKEN",
				Message: "Invalid user ID in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Generate new access token
	newToken, err := middleware.GenerateToken(
		user.ID.String(),
		user.Email,
		models.RoleUser,
		h.jwtConfig,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Failed to generate new access token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	response := gin.H{
		"access_token": newToken,
		"token_type":   "Bearer",
		"expires_in":   int64(h.jwtConfig.Expiration.Seconds()),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		Timestamp: time.Now(),
	})
}

// Logout handles user logout requests
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from context (set by JWT middleware)
	claims, exists := c.Get("jwt_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "No valid token found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	jwtClaims, ok := claims.(*middleware.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to process token claims",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Revoke user session (if token ID is available)
	// This is a simplified approach - in production you might want to maintain
	// a token blacklist or use a proper session management system
	sessionID, err := uuid.Parse(jwtClaims.ID)
	if err == nil {
		err = h.userService.RevokeUserSession(c.Request.Context(), sessionID)
	}
	if err != nil {
		// Log error but don't fail logout
		// User might be logging out with an already expired/invalid session
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "Successfully logged out",
		},
		Timestamp: time.Now(),
	})
}

// GetProfile returns the current user's profile (protected endpoint example)
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      user.ToResponse(),
		Timestamp: time.Now(),
	})
}

// Helper function to extract validation errors
func getValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = field + " must be a valid email address"
			case "min":
				errors[field] = field + " must be at least " + e.Param() + " characters long"
			case "max":
				errors[field] = field + " must be at most " + e.Param() + " characters long"
			default:
				errors[field] = field + " is invalid"
			}
		}
	} else {
		errors["general"] = err.Error()
	}
	
	return errors
}
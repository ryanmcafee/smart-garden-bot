package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

// GardenHandler handles garden-related requests
type GardenHandler struct {
	gardenService *services.GardenService
	validator     *validator.Validate
}

// NewGardenHandler creates a new garden handler
func NewGardenHandler(gardenService *services.GardenService) *GardenHandler {
	return &GardenHandler{
		gardenService: gardenService,
		validator:     validator.New(),
	}
}

// ===== GARDEN ENDPOINTS =====

// ListGardens lists all gardens for the authenticated user
func (h *GardenHandler) ListGardens(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardens, err := h.gardenService.ListGardens(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FETCH_FAILED",
				Message: "Failed to retrieve gardens",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      gardens,
		Timestamp: time.Now(),
	})
}

// CreateGarden creates a new garden
func (h *GardenHandler) CreateGarden(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	var req models.CreateGardenRequest
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

	garden, err := h.gardenService.CreateGarden(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATION_FAILED",
				Message: "Failed to create garden",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      garden,
		Timestamp: time.Now(),
	})
}

// GetGarden retrieves a specific garden
func (h *GardenHandler) GetGarden(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardenIDStr := c.Param("id")
	gardenID, err := uuid.Parse(gardenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_GARDEN_ID",
				Message: "Invalid garden ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	garden, err := h.gardenService.GetGarden(c.Request.Context(), userID, gardenID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "GARDEN_NOT_FOUND",
				Message: "Garden not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      garden,
		Timestamp: time.Now(),
	})
}

// UpdateGarden updates an existing garden
func (h *GardenHandler) UpdateGarden(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardenIDStr := c.Param("id")
	gardenID, err := uuid.Parse(gardenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_GARDEN_ID",
				Message: "Invalid garden ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	var req models.UpdateGardenRequest
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

	garden, err := h.gardenService.UpdateGarden(c.Request.Context(), userID, gardenID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update garden",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      garden,
		Timestamp: time.Now(),
	})
}

// DeleteGarden deletes a garden
func (h *GardenHandler) DeleteGarden(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardenIDStr := c.Param("id")
	gardenID, err := uuid.Parse(gardenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_GARDEN_ID",
				Message: "Invalid garden ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	err = h.gardenService.DeleteGarden(c.Request.Context(), userID, gardenID)
	if err != nil {
		if err.Error() == "garden not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "GARDEN_NOT_FOUND",
					Message: "Garden not found",
				},
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete garden",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ===== ZONE ENDPOINTS =====

// ListZones lists all zones for a garden
func (h *GardenHandler) ListZones(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardenIDStr := c.Param("id")
	gardenID, err := uuid.Parse(gardenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_GARDEN_ID",
				Message: "Invalid garden ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Verify user owns the garden
	_, err = h.gardenService.GetGarden(c.Request.Context(), userID, gardenID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "GARDEN_NOT_FOUND",
				Message: "Garden not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	zones, err := h.gardenService.ListZones(c.Request.Context(), gardenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FETCH_FAILED",
				Message: "Failed to retrieve zones",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      zones,
		Timestamp: time.Now(),
	})
}

// CreateZone creates a new zone
func (h *GardenHandler) CreateZone(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardenIDStr := c.Param("id")
	gardenID, err := uuid.Parse(gardenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_GARDEN_ID",
				Message: "Invalid garden ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Verify user owns the garden
	_, err = h.gardenService.GetGarden(c.Request.Context(), userID, gardenID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "GARDEN_NOT_FOUND",
				Message: "Garden not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateZoneRequest
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

	zone, err := h.gardenService.CreateZone(c.Request.Context(), gardenID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATION_FAILED",
				Message: "Failed to create zone",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      zone,
		Timestamp: time.Now(),
	})
}

// UpdateZone updates an existing zone
func (h *GardenHandler) UpdateZone(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardenIDStr := c.Param("id")
	gardenID, err := uuid.Parse(gardenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_GARDEN_ID",
				Message: "Invalid garden ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	zoneIDStr := c.Param("zone_id")
	zoneID, err := uuid.Parse(zoneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ZONE_ID",
				Message: "Invalid zone ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Verify user owns the garden
	_, err = h.gardenService.GetGarden(c.Request.Context(), userID, gardenID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "GARDEN_NOT_FOUND",
				Message: "Garden not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	var req models.UpdateZoneRequest
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

	zone, err := h.gardenService.UpdateZone(c.Request.Context(), gardenID, zoneID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update zone",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      zone,
		Timestamp: time.Now(),
	})
}

// DeleteZone deletes a zone
func (h *GardenHandler) DeleteZone(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
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

	gardenIDStr := c.Param("id")
	gardenID, err := uuid.Parse(gardenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_GARDEN_ID",
				Message: "Invalid garden ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	zoneIDStr := c.Param("zone_id")
	zoneID, err := uuid.Parse(zoneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ZONE_ID",
				Message: "Invalid zone ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Verify user owns the garden
	_, err = h.gardenService.GetGarden(c.Request.Context(), userID, gardenID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "GARDEN_NOT_FOUND",
				Message: "Garden not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	err = h.gardenService.DeleteZone(c.Request.Context(), gardenID, zoneID)
	if err != nil {
		if err.Error() == "zone not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "ZONE_NOT_FOUND",
					Message: "Zone not found",
				},
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete zone",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
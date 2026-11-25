package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

// WateringHandler handles watering-related requests
type WateringHandler struct {
	gardenService *services.GardenService
	sensorService *services.SensorService
	validator     *validator.Validate
}

// NewWateringHandler creates a new watering handler
func NewWateringHandler(gardenService *services.GardenService) *WateringHandler {
	return &WateringHandler{
		gardenService: gardenService,
		validator:     validator.New(),
	}
}

// GetSchedules retrieves watering schedules for a garden or zone
func (h *WateringHandler) GetSchedules(c *gin.Context) {
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

	gardenIDStr := c.Query("garden_id")
	zoneIDStr := c.Query("zone_id")

	if gardenIDStr == "" && zoneIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_PARAMETERS",
				Message: "Either garden_id or zone_id is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// TODO: Implement watering schedule retrieval
	// This would require adding schedule methods to the garden service
	// For now, return a placeholder response

	schedules := []models.WateringSchedule{}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      schedules,
		Timestamp: time.Now(),
	})
}

// CreateSchedule creates a new watering schedule
func (h *WateringHandler) CreateSchedule(c *gin.Context) {
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

	zoneIDStr := c.Query("zone_id")
	if zoneIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_ZONE_ID",
				Message: "Zone ID is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

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

	var req models.CreateWateringScheduleRequest
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

	// TODO: Verify user owns the zone
	// TODO: Create watering schedule

	// Placeholder response
	schedule := models.WateringSchedule{
		ID:              uuid.New(),
		ZoneID:          zoneID,
		Name:            req.Name,
		CronExpression:  req.CronExpression,
		DurationMinutes: req.DurationMinutes,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      schedule,
		Timestamp: time.Now(),
	})
}

// UpdateSchedule updates an existing watering schedule
func (h *WateringHandler) UpdateSchedule(c *gin.Context) {
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

	scheduleIDStr := c.Param("id")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_SCHEDULE_ID",
				Message: "Invalid schedule ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	var req models.UpdateWateringScheduleRequest
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

	// TODO: Verify user owns the schedule
	// TODO: Update watering schedule

	// Placeholder response
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"message":     "Schedule updated",
			"schedule_id": scheduleID,
		},
		Timestamp: time.Now(),
	})
}

// DeleteSchedule deletes a watering schedule
func (h *WateringHandler) DeleteSchedule(c *gin.Context) {
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

	scheduleIDStr := c.Param("id")
	_, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_SCHEDULE_ID",
				Message: "Invalid schedule ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// TODO: Verify user owns the schedule
	// TODO: Delete watering schedule

	c.JSON(http.StatusNoContent, nil)
}

// ManualWatering triggers manual watering for a zone
func (h *WateringHandler) ManualWatering(c *gin.Context) {
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

	var req models.ManualWateringRequest
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

	// TODO: Verify user owns the zone
	// TODO: Trigger manual watering

	// Create a watering event if sensor service is available
	if h.sensorService != nil {
		reason := "Manual watering triggered by user"
		if req.Reason != nil {
			reason = *req.Reason
		}

		wateringReq := &models.CreateWateringEventRequest{
			ZoneID:                 zoneID,
			PlannedDurationMinutes: req.DurationMinutes,
			TriggerType:            models.TriggerTypeManual,
			TriggerReason:          &reason,
		}

		event, err := h.sensorService.CreateWateringEvent(c.Request.Context(), wateringReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "WATERING_START_FAILED",
					Message: "Failed to start manual watering",
					Details: map[string]string{"error": err.Error()},
				},
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusOK, models.APIResponse{
			Success:   true,
			Data:      event,
			Timestamp: time.Now(),
		})
		return
	}

	// Fallback response if sensor service is not available
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"message":          "Manual watering initiated",
			"zone_id":          zoneID,
			"duration_minutes": req.DurationMinutes,
			"started_at":       time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	})
}

// GetWateringEvents retrieves watering events/history
func (h *WateringHandler) GetWateringEvents(c *gin.Context) {
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

	// Parse query parameters
	zoneIDStr := c.Query("zone_id")
	gardenIDStr := c.Query("garden_id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var zoneID, gardenID *uuid.UUID

	if zoneIDStr != "" {
		id, err := uuid.Parse(zoneIDStr)
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
		zoneID = &id
	}

	if gardenIDStr != "" {
		id, err := uuid.Parse(gardenIDStr)
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
		gardenID = &id
		
		// TODO: Verify user owns the garden
	}

	if zoneID == nil && gardenID == nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_PARAMETERS",
				Message: "Either zone_id or garden_id is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Get watering events if sensor service is available
	if h.sensorService != nil {
		events, err := h.sensorService.GetWateringEvents(c.Request.Context(), zoneID, gardenID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "EVENTS_FETCH_FAILED",
					Message: "Failed to retrieve watering events",
					Details: map[string]string{"error": err.Error()},
				},
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusOK, models.APIResponse{
			Success:   true,
			Data:      events,
			Timestamp: time.Now(),
		})
		return
	}

	// Fallback response if sensor service is not available
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      []models.WateringEvent{},
		Timestamp: time.Now(),
	})
}

// GetWateringStatus returns current watering status for zones
func (h *WateringHandler) GetWateringStatus(c *gin.Context) {
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

	gardenIDStr := c.Query("garden_id")
	if gardenIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_GARDEN_ID",
				Message: "Garden ID is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

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

	// TODO: Verify user owns the garden
	// TODO: Get current watering status for all zones in the garden

	// Placeholder response
	status := gin.H{
		"garden_id": gardenID,
		"zones":     []gin.H{},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      status,
		Timestamp: time.Now(),
	})
}
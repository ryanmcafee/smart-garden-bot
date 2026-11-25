package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

// SensorHandler handles sensor-related requests
type SensorHandler struct {
	sensorService *services.SensorService
	gardenService *services.GardenService
	validator     *validator.Validate
}

// NewSensorHandler creates a new sensor handler
func NewSensorHandler(sensorService *services.SensorService) *SensorHandler {
	return &SensorHandler{
		sensorService: sensorService,
		validator:     validator.New(),
	}
}

// GetReadings retrieves sensor readings with filtering
func (h *SensorHandler) GetReadings(c *gin.Context) {
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
	var query models.SensorReadingsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_QUERY_PARAMS",
				Message: "Invalid query parameters",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Set defaults
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.OrderBy == "" {
		query.OrderBy = "timestamp"
	}
	if query.OrderDir == "" {
		query.OrderDir = "desc"
	}

	// Validate request
	if err := h.validator.Struct(&query); err != nil {
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

	// TODO: Add authorization check - user should only access their own gardens' sensor data
	// This would require checking if the device_id, zone_id, or garden_id belongs to the user

	readings, err := h.sensorService.GetSensorReadings(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FETCH_FAILED",
				Message: "Failed to retrieve sensor readings",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      readings,
		Timestamp: time.Now(),
	})
}

// CreateReading creates a new sensor reading
func (h *SensorHandler) CreateReading(c *gin.Context) {
	// This endpoint can be used by IoT devices with API keys
	// or by authenticated users for manual data entry

	var req models.CreateSensorReadingRequest
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

	// TODO: Add authorization check - ensure device belongs to the user
	// or API key has the right permissions

	reading, err := h.sensorService.CreateSensorReading(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATION_FAILED",
				Message: "Failed to create sensor reading",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      reading,
		Timestamp: time.Now(),
	})
}

// GetTimeSeries retrieves time-series data for a specific metric
func (h *SensorHandler) GetTimeSeries(c *gin.Context) {
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
	deviceIDStr := c.Query("device_id")
	if deviceIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_DEVICE_ID",
				Message: "Device ID is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_DEVICE_ID",
				Message: "Invalid device ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	metric := c.Query("metric")
	if metric == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_METRIC",
				Message: "Metric parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Parse time range
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_START_TIME",
					Message: "Invalid start_time format, use RFC3339",
				},
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_END_TIME",
					Message: "Invalid end_time format, use RFC3339",
				},
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		endTime = time.Now()
	}

	// TODO: Add authorization check - ensure device belongs to the user

	timeSeriesData, err := h.sensorService.GetTimeSeriesData(c.Request.Context(), deviceID, metric, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FETCH_FAILED",
				Message: "Failed to retrieve time series data",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      timeSeriesData,
		Timestamp: time.Now(),
	})
}

// ListDevices lists all devices for the user's gardens
func (h *SensorHandler) ListDevices(c *gin.Context) {
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

	// Get garden ID from query params (optional)
	gardenIDStr := c.Query("garden_id")
	var devices []models.Device

	if gardenIDStr != "" {
		_, err := uuid.Parse(gardenIDStr)
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

		// TODO: Verify user owns the garden and get devices
		// This requires access to garden service
		devices = []models.Device{} // Placeholder
	} else {
		// Get devices for all user's gardens
		// TODO: Implement this by getting all user's gardens and their devices
		devices = []models.Device{} // Placeholder
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      devices,
		Timestamp: time.Now(),
	})
}

// RegisterDevice registers a new IoT device
func (h *SensorHandler) RegisterDevice(c *gin.Context) {
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

	var req models.CreateDeviceRequest
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

	// Get garden ID from the request
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

	_, err := uuid.Parse(gardenIDStr)
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

	// TODO: Verify user owns the garden and create device
	// This requires access to garden service
	// device, err := h.gardenService.CreateDevice(c.Request.Context(), gardenID, &req)

	// Placeholder response
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "Device registration endpoint not fully implemented",
		},
		Timestamp: time.Now(),
	})
}

// GetSensorStats returns sensor statistics
func (h *SensorHandler) GetSensorStats(c *gin.Context) {
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

	// Get garden ID from query params (optional)
	var gardenID *uuid.UUID
	gardenIDStr := c.Query("garden_id")
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

	stats, err := h.sensorService.GetSensorStats(c.Request.Context(), gardenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "STATS_FETCH_FAILED",
				Message: "Failed to retrieve sensor statistics",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      stats,
		Timestamp: time.Now(),
	})
}

// Helper function to parse time from query parameter
func parseTimeQuery(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return &t, nil
	}

	// Try Unix timestamp
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		t := time.Unix(timestamp, 0)
		return &t, nil
	}

	return nil, fmt.Errorf("invalid time format")
}

// Helper function to parse duration from query parameter
func parseDurationQuery(durationStr string) (*time.Duration, error) {
	if durationStr == "" {
		return nil, nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, err
	}

	return &duration, nil
}
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

// WeatherHandler handles weather-related requests
type WeatherHandler struct {
	weatherService *services.WeatherService
}

// NewWeatherHandler creates a new weather handler
func NewWeatherHandler(weatherService *services.WeatherService) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
	}
}

// GetCurrentWeather retrieves current weather conditions
func (h *WeatherHandler) GetCurrentWeather(c *gin.Context) {
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

	location := c.Query("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_LOCATION",
				Message: "Location parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	weather, err := h.weatherService.GetCurrentWeather(c.Request.Context(), location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEATHER_FETCH_FAILED",
				Message: "Failed to retrieve current weather",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Convert to standard response format
	currentWeather := models.CurrentWeatherData{
		Location:           weather.Location,
		Temperature:        weather.Temperature,
		Humidity:           weather.Humidity,
		Pressure:           weather.Pressure,
		WindSpeed:          weather.WindSpeed,
		WindDirection:      int(weather.WindDir),
		Rainfall:           weather.Rainfall,
		UVIndex:            &weather.UV,
		WeatherCode:        "clear", // Placeholder
		WeatherDescription: "Current conditions",
		Timestamp:          weather.Timestamp,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      currentWeather,
		Timestamp: time.Now(),
	})
}

// GetForecast retrieves weather forecast
func (h *WeatherHandler) GetForecast(c *gin.Context) {
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

	location := c.Query("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_LOCATION",
				Message: "Location parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Parse days parameter
	daysStr := c.DefaultQuery("days", "5")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 14 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_DAYS_PARAM",
				Message: "Days parameter must be between 1 and 14",
			},
			Timestamp: time.Now(),
		})
		return
	}

	forecast, err := h.weatherService.GetForecast(c.Request.Context(), location, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORECAST_FETCH_FAILED",
				Message: "Failed to retrieve weather forecast",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Convert to standard response format
	dailyForecasts := make([]models.DailyForecast, 0, len(forecast.Days))
	for _, day := range forecast.Days {
		dailyForecasts = append(dailyForecasts, models.DailyForecast{
			Date:               day.Date,
			TemperatureMax:     day.TempMax,
			TemperatureMin:     day.TempMin,
			Humidity:           day.Humidity,
			WindSpeed:          day.WindSpeed,
			WindDirection:      0, // Not available in current format
			RainProbability:    int(day.PrecipChance),
			RainfallTotal:      day.PrecipAmount,
			WeatherCode:        "forecast",
			WeatherDescription: day.Description,
		})
	}

	weatherForecast := models.WeatherForecast{
		Location:  forecast.Location,
		Timestamp: time.Now(),
		Days:      dailyForecasts,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      weatherForecast,
		Timestamp: time.Now(),
	})
}

// GetWateringRecommendation provides watering recommendations based on weather
func (h *WeatherHandler) GetWateringRecommendation(c *gin.Context) {
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

	location := c.Query("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_LOCATION",
				Message: "Location parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	shouldWater, reason, err := h.weatherService.GetWateringRecommendation(c.Request.Context(), location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "RECOMMENDATION_FAILED",
				Message: "Failed to get watering recommendation",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	recommendation := gin.H{
		"should_water": shouldWater,
		"reason":       reason,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      recommendation,
		Timestamp: time.Now(),
	})
}
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/database"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
)

// HealthCheck returns the basic health status
func HealthCheck(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]models.HealthCheck)
		overallStatus := "healthy"

		// Database health check
		start := time.Now()
		dbErr := database.HealthCheck(ctx, db)
		dbLatency := time.Since(start)

		if dbErr != nil {
			checks["database"] = models.HealthCheck{
				Status:  "unhealthy",
				Message: dbErr.Error(),
				Latency: dbLatency,
			}
			overallStatus = "unhealthy"
		} else {
			checks["database"] = models.HealthCheck{
				Status:  "healthy",
				Latency: dbLatency,
			}
		}

		// Database pool stats
		stats := database.GetStats(db)
		openConns := stats.OpenConnections
		if openConns == 0 {
			checks["database_pool"] = models.HealthCheck{
				Status:  "warning",
				Message: "No active database connections",
			}
		} else {
			checks["database_pool"] = models.HealthCheck{
				Status: "healthy",
				Message: fmt.Sprintf("Open: %d, InUse: %d, Idle: %d",
					stats.OpenConnections,
					stats.InUse,
					stats.Idle),
			}
		}

		response := models.HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now(),
			Version:   getVersion(),
			Checks:    checks,
		}

		statusCode := http.StatusOK
		if overallStatus == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, response)
	}
}

// ReadinessCheck returns the readiness status (similar to health but for K8s)
func ReadinessCheck(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Quick database ping
		if err := database.HealthCheck(ctx, db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not ready",
				"message": "Database connection failed",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now(),
		})
	}
}

// LivenessCheck returns a simple liveness probe
func LivenessCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now(),
		})
	}
}

// PrometheusMetrics returns the Prometheus metrics endpoint
func PrometheusMetrics() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// getVersion returns the application version
func getVersion() string {
	// In a real application, this would be set during build
	return "1.0.0"
}
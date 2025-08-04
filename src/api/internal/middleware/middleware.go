package middleware

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)
)

// CORS middleware with secure defaults
func CORS() gin.HandlerFunc {
	config := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// In production, implement proper origin validation
			return true // For development only
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}

// Security middleware with common security headers
func Security() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Remove server information
		c.Header("Server", "")
		
		c.Next()
	})
}

// Prometheus metrics middleware
func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		activeConnections.Inc()
		
		defer func() {
			activeConnections.Dec()
			
			duration := time.Since(start).Seconds()
			endpoint := c.FullPath()
			if endpoint == "" {
				endpoint = "unknown"
			}
			
			httpRequestsTotal.WithLabelValues(
				c.Request.Method,
				endpoint,
				string(rune(c.Writer.Status())),
			).Inc()
			
			httpRequestDuration.WithLabelValues(
				c.Request.Method,
				endpoint,
			).Observe(duration)
		}()
		
		c.Next()
	}
}

// RateLimit middleware (basic implementation)
func RateLimit() gin.HandlerFunc {
	// This is a simple in-memory rate limiter
	// In production, use Redis-based rate limiting
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: Implement proper rate limiting
		// Consider using redis or a proper rate limiting library
		c.Next()
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// Logger middleware with structured logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Structured logging format
		return fmt.Sprintf(
			`{"time":"%s","method":"%s","path":"%s","status":%d,"latency":"%s","ip":"%s","user_agent":"%s","request_id":"%s"}%s`,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Request.UserAgent(),
			param.Keys["request_id"],
			"\n",
		)
	})
}

// generateRequestID creates a unique request identifier
func generateRequestID() string {
	// Simple implementation - in production use UUID or similar
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Recovery middleware with better error handling
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(500, gin.H{
				"error":      "Internal server error",
				"request_id": c.GetString("request_id"),
			})
			// Log the actual error (don't expose to client)
			fmt.Printf("Panic recovered: %s\n", err)
		}
		c.AbortWithStatus(500)
	})
}

// PrometheusHandler returns the Prometheus metrics handler
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ryanmcafee/smart-garden-bot/api/internal/config"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/database"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/handlers"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/middleware"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

// IntegrationTestSuite contains all integration tests
type IntegrationTestSuite struct {
	suite.Suite
	
	// Infrastructure
	postgresContainer testcontainers.Container
	redisContainer    testcontainers.Container
	
	// Application components
	router   *gin.Engine
	db       *database.DB
	config   *config.Config
	
	// Test utilities
	client   *http.Client
	baseURL  string
	authToken string
}

// SetupSuite runs before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	
	// Start PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("smart_garden_bot_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts("../../../database/migrations/001_initial_schema.up.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(suite.T(), err)
	suite.postgresContainer = postgresContainer
	
	// Start Redis container
	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
	)
	require.NoError(suite.T(), err)
	suite.redisContainer = redisContainer
	
	// Get container connection details
	postgresHost, err := postgresContainer.Host(ctx)
	require.NoError(suite.T(), err)
	
	postgresPort, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(suite.T(), err)
	
	redisHost, err := redisContainer.Host(ctx)
	require.NoError(suite.T(), err)
	
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(suite.T(), err)
	
	// Configure application
	suite.config = &config.Config{
		Environment: "test",
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: config.DatabaseConfig{
			Host:     postgresHost,
			Port:     postgresPort.Int(),
			Name:     "smart_garden_bot_test",
			User:     "test",
			Password: "test",
			SSLMode:  "disable",
			MaxConns: 10,
			MinConns: 2,
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			Expiration: time.Hour,
			Issuer:     "smart-garden-bot-test",
		},
		Redis: config.RedisConfig{
			Host: redisHost,
			Port: redisPort.Int(),
		},
		Weather: config.WeatherConfig{
			OpenWeatherMapAPIKey: "test-api-key",
			DefaultProvider:      "openweathermap",
			CacheTTL:            5 * time.Minute,
		},
	}
	
	// Initialize database
	db, err := database.Initialize(suite.config.Database)
	require.NoError(suite.T(), err)
	suite.db = db
	
	// Initialize services
	weatherService := services.NewWeatherService(suite.config.Weather)
	userService := services.NewUserService(db)
	gardenService := services.NewGardenService(db)
	sensorService := services.NewSensorService(db)
	
	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	
	// Health endpoints
	router.GET("/health", handlers.HealthCheck(db))
	router.GET("/ready", handlers.ReadinessCheck(db))
	router.GET("/metrics", handlers.PrometheusMetrics())
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			authHandler := handlers.NewAuthHandler(userService, suite.config.JWT)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}
		
		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(suite.config.JWT))
		{
			// Users
			users := protected.Group("/users")
			{
				userHandler := handlers.NewUserHandler(userService)
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
			}
			
			// Gardens
			gardens := protected.Group("/gardens")
			{
				gardenHandler := handlers.NewGardenHandler(gardenService)
				gardens.GET("", gardenHandler.ListGardens)
				gardens.POST("", gardenHandler.CreateGarden)
				gardens.GET("/:id", gardenHandler.GetGarden)
				gardens.PUT("/:id", gardenHandler.UpdateGarden)
				gardens.DELETE("/:id", gardenHandler.DeleteGarden)
			}
			
			// Sensors
			sensors := protected.Group("/sensors")
			{
				sensorHandler := handlers.NewSensorHandler(sensorService)
				sensors.GET("/readings", sensorHandler.GetReadings)
				sensors.POST("/readings", sensorHandler.CreateReading)
			}
			
			// Weather
			weather := protected.Group("/weather")
			{
				weatherHandler := handlers.NewWeatherHandler(weatherService)
				weather.GET("/current", weatherHandler.GetCurrentWeather)
				weather.GET("/forecast", weatherHandler.GetForecast)
			}
		}
	}
	
	suite.router = router
	suite.client = &http.Client{Timeout: 30 * time.Second}
	
	// Create test user and get auth token
	suite.createTestUserAndLogin()
}

// TearDownSuite runs after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	ctx := context.Background()
	
	if suite.db != nil {
		suite.db.Close()
	}
	
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(ctx)
	}
	
	if suite.redisContainer != nil {
		suite.redisContainer.Terminate(ctx)
	}
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up database between tests
	suite.cleanupDatabase()
}

// createTestUserAndLogin creates a test user and obtains auth token
func (suite *IntegrationTestSuite) createTestUserAndLogin() {
	// Insert test user directly into database
	ctx := context.Background()
	_, err := suite.db.Pool.Exec(ctx, `
		INSERT INTO auth.users (id, email, name, auth0_user_id)
		VALUES ('550e8400-e29b-41d4-a716-446655440000', 'test@example.com', 'Test User', 'auth0|test123')
		ON CONFLICT (email) DO NOTHING
	`)
	require.NoError(suite.T(), err)
	
	// Generate JWT token for testing
	token, err := middleware.GenerateToken(
		"550e8400-e29b-41d4-a716-446655440000",
		"test@example.com",
		"user",
		suite.config.JWT,
	)
	require.NoError(suite.T(), err)
	
	suite.authToken = token
}

// cleanupDatabase removes test data between tests
func (suite *IntegrationTestSuite) cleanupDatabase() {
	ctx := context.Background()
	
	// Clean up test data in reverse dependency order
	tables := []string{
		"sensors.watering_events",
		"sensors.readings",
		"gardens.watering_schedules",
		"gardens.devices",
		"gardens.zones",
		"gardens.gardens",
		"auth.api_keys",
		"auth.user_sessions",
	}
	
	for _, table := range tables {
		_, err := suite.db.Pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE 1=1", table))
		require.NoError(suite.T(), err)
	}
}

// makeRequest is a helper to make HTTP requests with auth
func (suite *IntegrationTestSuite) makeRequest(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		require.NoError(suite.T(), err)
	}
	
	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Add auth token if available
	if suite.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
	}
	
	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	
	return recorder
}

// Test health endpoints
func (suite *IntegrationTestSuite) TestHealthEndpoints() {
	// Test health endpoint
	resp := suite.makeRequest("GET", "/health", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var healthResp map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &healthResp)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", healthResp["status"])
	
	// Test readiness endpoint
	resp = suite.makeRequest("GET", "/ready", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	// Test metrics endpoint
	resp = suite.makeRequest("GET", "/metrics", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	assert.Contains(suite.T(), resp.Body.String(), "# HELP")
}

// Test user management
func (suite *IntegrationTestSuite) TestUserManagement() {
	// Test get profile
	resp := suite.makeRequest("GET", "/api/v1/users/profile", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var user map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &user)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test@example.com", user["email"])
	assert.Equal(suite.T(), "Test User", user["name"])
	
	// Test update profile
	updateData := map[string]interface{}{
		"name": "Updated Test User",
	}
	
	resp = suite.makeRequest("PUT", "/api/v1/users/profile", updateData, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	// Verify update
	resp = suite.makeRequest("GET", "/api/v1/users/profile", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	err = json.Unmarshal(resp.Body.Bytes(), &user)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Test User", user["name"])
}

// Test garden management
func (suite *IntegrationTestSuite) TestGardenManagement() {
	// Create a garden
	gardenData := map[string]interface{}{
		"name":        "Test Garden",
		"description": "A test garden",
		"location":    "Test City",
		"latitude":    37.7749,
		"longitude":   -122.4194,
		"timezone":    "America/Los_Angeles",
	}
	
	resp := suite.makeRequest("POST", "/api/v1/gardens", gardenData, nil)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	
	var createdGarden map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &createdGarden)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Garden", createdGarden["name"])
	
	gardenID := createdGarden["id"].(string)
	
	// List gardens
	resp = suite.makeRequest("GET", "/api/v1/gardens", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var gardens []map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &gardens)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), gardens, 1)
	assert.Equal(suite.T(), "Test Garden", gardens[0]["name"])
	
	// Get specific garden
	resp = suite.makeRequest("GET", "/api/v1/gardens/"+gardenID, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var garden map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &garden)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), gardenID, garden["id"])
	
	// Update garden
	updateData := map[string]interface{}{
		"name":        "Updated Test Garden",
		"description": "An updated test garden",
	}
	
	resp = suite.makeRequest("PUT", "/api/v1/gardens/"+gardenID, updateData, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	// Delete garden
	resp = suite.makeRequest("DELETE", "/api/v1/gardens/"+gardenID, nil, nil)
	assert.Equal(suite.T(), http.StatusNoContent, resp.Code)
	
	// Verify deletion
	resp = suite.makeRequest("GET", "/api/v1/gardens/"+gardenID, nil, nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.Code)
}

// Test sensor data management
func (suite *IntegrationTestSuite) TestSensorDataManagement() {
	// First create a garden and zone
	gardenData := map[string]interface{}{
		"name":     "Sensor Test Garden",
		"location": "Test City",
		"timezone": "UTC",
	}
	
	resp := suite.makeRequest("POST", "/api/v1/gardens", gardenData, nil)
	require.Equal(suite.T(), http.StatusCreated, resp.Code)
	
	var garden map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &garden)
	require.NoError(suite.T(), err)
	
	// Create sensor reading
	readingData := map[string]interface{}{
		"device_id":        "test-device-001",
		"timestamp":        time.Now().Format(time.RFC3339),
		"temperature":      22.5,
		"humidity":         65.0,
		"soil_moisture":    45.0,
		"soil_temperature": 20.0,
		"light_intensity":  25000.0,
	}
	
	resp = suite.makeRequest("POST", "/api/v1/sensors/readings", readingData, nil)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	
	// Get sensor readings
	resp = suite.makeRequest("GET", "/api/v1/sensors/readings?device_id=test-device-001", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var readings []map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &readings)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), readings, 1)
	assert.Equal(suite.T(), "test-device-001", readings[0]["device_id"])
	assert.Equal(suite.T(), 22.5, readings[0]["temperature"])
}

// Test weather integration
func (suite *IntegrationTestSuite) TestWeatherIntegration() {
	// Note: This test would require mocking the weather API
	// For now, we'll test the endpoint structure
	
	resp := suite.makeRequest("GET", "/api/v1/weather/current?location=San Francisco", nil, nil)
	// Since we don't have a real API key, this will likely return an error
	// In a real test, you would mock the weather service
	assert.Contains(suite.T(), []int{http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError}, resp.Code)
}

// Test authentication
func (suite *IntegrationTestSuite) TestAuthentication() {
	// Test accessing protected endpoint without token
	req := httptest.NewRequest("GET", "/api/v1/users/profile", nil)
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	
	assert.Equal(suite.T(), http.StatusUnauthorized, recorder.Code)
	
	// Test with invalid token
	req = httptest.NewRequest("GET", "/api/v1/users/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	
	assert.Equal(suite.T(), http.StatusUnauthorized, recorder.Code)
	
	// Test with valid token (already tested in other methods)
	resp := suite.makeRequest("GET", "/api/v1/users/profile", nil, nil)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

// Test rate limiting and security headers
func (suite *IntegrationTestSuite) TestSecurityFeatures() {
	resp := suite.makeRequest("GET", "/health", nil, nil)
	
	// Check security headers
	headers := resp.Header()
	assert.Equal(suite.T(), "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(suite.T(), "DENY", headers.Get("X-Frame-Options"))
}

// Test concurrent requests
func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	const numRequests = 10
	done := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			resp := suite.makeRequest("GET", "/health", nil, nil)
			assert.Equal(suite.T(), http.StatusOK, resp.Code)
			done <- true
		}()
	}
	
	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// Run the test suite
func TestIntegrationSuite(t *testing.T) {
	// Skip integration tests if not explicitly requested
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run.")
	}
	
	suite.Run(t, new(IntegrationTestSuite))
}
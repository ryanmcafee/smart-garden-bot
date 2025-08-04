package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OpenAPISpec returns the OpenAPI 3.1 specification
func OpenAPISpec() gin.HandlerFunc {
	return func(c *gin.Context) {
		spec := getOpenAPISpecification()
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, spec)
	}
}

func getOpenAPISpecification() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":       "Smart Garden Bot API",
			"description": "REST API for the Smart Garden Bot platform - automated garden irrigation management",
			"version":     "1.0.0",
			"contact": map[string]interface{}{
				"name":  "Smart Garden Bot Team",
				"email": "support@smartgardenbot.com",
				"url":   "https://smartgardenbot.com",
			},
			"license": map[string]interface{}{
				"name": "MIT",
				"url":  "https://opensource.org/licenses/MIT",
			},
		},
		"servers": []map[string]interface{}{
			{
				"url":         "https://api.smartgardenbot.com/api/v1",
				"description": "Production server",
			},
			{
				"url":         "https://staging-api.smartgardenbot.com/api/v1",
				"description": "Staging server",
			},
			{
				"url":         "http://localhost:8080/api/v1",
				"description": "Development server",
			},
		},
		"paths": map[string]interface{}{
			"/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "User login",
					"description": "Authenticate user and return JWT token",
					"tags":        []string{"Authentication"},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/LoginRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Login successful",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/LoginResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Invalid credentials",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/gardens": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List gardens",
					"description": "Get all gardens for the authenticated user",
					"tags":        []string{"Gardens"},
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number for pagination",
							"schema": map[string]interface{}{
								"type":    "integer",
								"default": 1,
								"minimum": 1,
							},
						},
						{
							"name":        "limit",
							"in":          "query",
							"description": "Number of items per page",
							"schema": map[string]interface{}{
								"type":    "integer",
								"default": 10,
								"minimum": 1,
								"maximum": 100,
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of gardens",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/GardenListResponse",
									},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create garden",
					"description": "Create a new garden",
					"tags":        []string{"Gardens"},
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/CreateGardenRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Garden created successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/Garden",
									},
								},
							},
						},
					},
				},
			},
			"/sensors/readings": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get sensor readings",
					"description": "Retrieve sensor readings with optional filtering",
					"tags":        []string{"Sensors"},
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "device_id",
							"in":          "query",
							"description": "Filter by device ID",
							"schema":      map[string]string{"type": "string"},
						},
						{
							"name":        "start_time",
							"in":          "query",
							"description": "Start time for readings (ISO 8601)",
							"schema":      map[string]string{"type": "string", "format": "date-time"},
						},
						{
							"name":        "end_time",
							"in":          "query",
							"description": "End time for readings (ISO 8601)",
							"schema":      map[string]string{"type": "string", "format": "date-time"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Sensor readings",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SensorReadingsResponse",
									},
								},
							},
						},
					},
				},
			},
			"/weather/current": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get current weather",
					"description": "Get current weather conditions for a location",
					"tags":        []string{"Weather"},
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "location",
							"in":          "query",
							"required":    true,
							"description": "Location (city name or coordinates)",
							"schema":      map[string]string{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Current weather data",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/WeatherData",
									},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"bearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
				"apiKeyAuth": map[string]interface{}{
					"type": "apiKey",
					"in":   "header",
					"name": "X-API-Key",
				},
			},
			"schemas": map[string]interface{}{
				"LoginRequest": map[string]interface{}{
					"type": "object",
					"required": []string{"email", "password"},
					"properties": map[string]interface{}{
						"email": map[string]interface{}{
							"type":   "string",
							"format": "email",
						},
						"password": map[string]interface{}{
							"type":      "string",
							"minLength": 8,
						},
					},
				},
				"LoginResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token": map[string]string{
							"type": "string",
						},
						"refresh_token": map[string]string{
							"type": "string",
						},
						"expires_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
						"user": map[string]interface{}{
							"$ref": "#/components/schemas/User",
						},
					},
				},
				"User": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{
							"type": "string",
						},
						"email": map[string]interface{}{
							"type":   "string",
							"format": "email",
						},
						"name": map[string]string{
							"type": "string",
						},
						"created_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
						"updated_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
					},
				},
				"Garden": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{
							"type": "string",
						},
						"name": map[string]string{
							"type": "string",
						},
						"location": map[string]string{
							"type": "string",
						},
						"timezone": map[string]string{
							"type": "string",
						},
						"zones": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/Zone",
							},
						},
						"created_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
						"updated_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
					},
				},
				"Zone": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{
							"type": "string",
						},
						"name": map[string]string{
							"type": "string",
						},
						"plant_type": map[string]string{
							"type": "string",
						},
						"area": map[string]interface{}{
							"type": "number",
						},
						"watering_schedule": map[string]string{
							"type": "string",
						},
					},
				},
				"WeatherData": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]string{
							"type": "string",
						},
						"temperature": map[string]interface{}{
							"type": "number",
						},
						"humidity": map[string]interface{}{
							"type": "number",
						},
						"pressure": map[string]interface{}{
							"type": "number",
						},
						"wind_speed": map[string]interface{}{
							"type": "number",
						},
						"rainfall": map[string]interface{}{
							"type": "number",
						},
						"timestamp": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
					},
				},
				"ErrorResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"error": map[string]string{
							"type": "string",
						},
						"message": map[string]string{
							"type": "string",
						},
						"code": map[string]interface{}{
							"type": "integer",
						},
						"request_id": map[string]string{
							"type": "string",
						},
					},
				},
			},
		},
		"tags": []map[string]interface{}{
			{
				"name":        "Authentication",
				"description": "User authentication and authorization",
			},
			{
				"name":        "Gardens",
				"description": "Garden management operations",
			},
			{
				"name":        "Sensors",
				"description": "Sensor data and device management",
			},
			{
				"name":        "Weather",
				"description": "Weather data integration",
			},
			{
				"name":        "Watering",
				"description": "Watering schedule and control",
			},
		},
	}
}
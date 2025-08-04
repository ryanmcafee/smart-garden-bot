package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/config"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/database"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/handlers"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/middleware"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize services
	weatherService := services.NewWeatherService(cfg.Weather)
	userService := services.NewUserService(db)
	gardenService := services.NewGardenService(db)
	sensorService := services.NewSensorService(db)
	billingService := services.NewBillingService(db, cfg.Stripe.APIKey)

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Security())
	router.Use(middleware.Prometheus())

	// Health and monitoring endpoints
	router.GET("/health", handlers.HealthCheck(db))
	router.GET("/ready", handlers.ReadinessCheck(db))
	router.GET("/metrics", handlers.PrometheusMetrics())
	router.GET("/liveness", handlers.LivenessCheck())

	// OpenAPI and documentation
	router.GET("/openapi.json", handlers.OpenAPISpec())
	router.Static("/swagger-ui", "./static/swagger-ui")

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			authHandler := handlers.NewAuthHandler(userService, cfg.JWT)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.JWTAuth(cfg.JWT), authHandler.Logout)
		}

		// Stripe webhook (not authenticated)
		v1.POST("/billing/webhook/stripe", handlers.NewBillingHandler(billingService).ProcessStripeWebhook)
		
		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(cfg.JWT))
		{
			// User management
			users := protected.Group("/users")
			{
				userHandler := handlers.NewUserHandler(userService)
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
				users.POST("/api-keys", userHandler.CreateAPIKey)
				users.GET("/api-keys", userHandler.ListAPIKeys)
				users.DELETE("/api-keys/:id", userHandler.RevokeAPIKey)
			}

			// Garden management
			gardens := protected.Group("/gardens")
			{
				gardenHandler := handlers.NewGardenHandler(gardenService)
				gardens.GET("", gardenHandler.ListGardens)
				gardens.POST("", gardenHandler.CreateGarden)
				gardens.GET("/:id", gardenHandler.GetGarden)
				gardens.PUT("/:id", gardenHandler.UpdateGarden)
				gardens.DELETE("/:id", gardenHandler.DeleteGarden)

				// Garden zones
				gardens.GET("/:id/zones", gardenHandler.ListZones)
				gardens.POST("/:id/zones", gardenHandler.CreateZone)
				gardens.PUT("/:id/zones/:zone_id", gardenHandler.UpdateZone)
				gardens.DELETE("/:id/zones/:zone_id", gardenHandler.DeleteZone)
			}

			// Sensor data
			sensors := protected.Group("/sensors")
			{
				sensorHandler := handlers.NewSensorHandler(sensorService)
				sensors.GET("/readings", sensorHandler.GetReadings)
				sensors.POST("/readings", sensorHandler.CreateReading)
				sensors.GET("/readings/timeseries", sensorHandler.GetTimeSeries)
				sensors.GET("/devices", sensorHandler.ListDevices)
				sensors.POST("/devices", sensorHandler.RegisterDevice)
			}

			// Weather integration
			weather := protected.Group("/weather")
			{
				weatherHandler := handlers.NewWeatherHandler(weatherService)
				weather.GET("/current", weatherHandler.GetCurrentWeather)
				weather.GET("/forecast", weatherHandler.GetForecast)
			}

			// Watering management
			watering := protected.Group("/watering")
			{
				wateringHandler := handlers.NewWateringHandler(gardenService)
				watering.GET("/schedules", wateringHandler.GetSchedules)
				watering.POST("/schedules", wateringHandler.CreateSchedule)
				watering.PUT("/schedules/:id", wateringHandler.UpdateSchedule)
				watering.DELETE("/schedules/:id", wateringHandler.DeleteSchedule)
				watering.POST("/manual/:zone_id", wateringHandler.ManualWatering)
			}

			// Billing and subscriptions
			billing := protected.Group("/billing")
			{
				billingHandler := handlers.NewBillingHandler(billingService)
				// Subscription plans
				billing.GET("/plans", billingHandler.GetSubscriptionPlans)
				billing.GET("/plans/:id", billingHandler.GetSubscriptionPlan)
				
				// Current subscription management
				billing.GET("/subscription", billingHandler.GetCurrentSubscription)
				billing.POST("/subscription", billingHandler.CreateSubscription)
				billing.PUT("/subscription", billingHandler.UpdateSubscription)
				billing.DELETE("/subscription", billingHandler.CancelSubscription)
				
				// Payment methods and invoices
				billing.GET("/payment-methods", billingHandler.GetPaymentMethods)
				billing.GET("/invoices", billingHandler.GetInvoices)
				
				// Usage and billing portal
				billing.GET("/usage", billingHandler.GetUsage)
				billing.POST("/portal", billingHandler.CreateBillingPortalSession)
				billing.GET("/preview-change", billingHandler.PreviewSubscriptionChange)
			}
			
			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
				// User statistics
				userHandler := handlers.NewUserHandler(userService)
				admin.GET("/users/stats", userHandler.GetUserStats)
				
				// Billing statistics
				billingHandler := handlers.NewBillingHandler(billingService)
				admin.GET("/billing/stats", billingHandler.GetSubscriptionStats)
			}
		}
	}

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

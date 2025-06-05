package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smarthome/telemetry/handlers"
	"smarthome/telemetry/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Get configuration from environment variables
	sensorsBaseURL := getEnv("SENSORS_API_URL", "http://smarthome-devices:8080/api/v1/sensors")
	temperatureBaseURL := getEnv("TEMPERATURE_API_URL", "http://temperature-api:8080/api/v2/temperature")
	port := getEnv("PORT", "8080")

	// Initialize services
	telemetryService := services.NewTelemetryService(sensorsBaseURL, temperatureBaseURL)

	// Initialize router
	router := gin.Default()
	api := router.Group("/api/v2")

	// Initialize handlers
	telemetryHandler := handlers.NewTelemetryHandler(telemetryService)
	telemetryHandler.RegisterRoutes(api)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting telemetry service on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

package main

import (
	"log"
	"net/http"

	"github.com/kcansari/task-management-api/config"
	"github.com/kcansari/task-management-api/database"
	"github.com/kcansari/task-management-api/handlers"
)

func main() {
	cfg := config.Load()

	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	if err := database.HealthCheck(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	// Root endpoint - simple welcome message
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World! Task Management API is running."))
	})

	// Health check endpoint - verifies database connectivity
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Authentication endpoints
	// POST /api/auth/register - Register a new user
	http.HandleFunc("/api/auth/register", handlers.Register)
	
	// POST /api/auth/login - Login existing user
	http.HandleFunc("/api/auth/login", handlers.Login)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

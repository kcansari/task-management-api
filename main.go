package main

import (
	"log"
	"net/http"

	"github.com/kcansari/task-management-api/config"
	"github.com/kcansari/task-management-api/database"
	"github.com/kcansari/task-management-api/handlers"
	"github.com/kcansari/task-management-api/middleware"
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

	// Authentication endpoints (public - no middleware required)
	// POST /api/auth/register - Register a new user
	http.HandleFunc("/api/auth/register", handlers.Register)
	
	// POST /api/auth/login - Login existing user
	http.HandleFunc("/api/auth/login", handlers.Login)

	// Protected Task endpoints (require authentication)
	// These routes use middleware.AuthMiddleware to ensure user is authenticated
	// The middleware extracts JWT token, validates it, and adds user info to context
	
	// Handle /api/tasks (without trailing slash) - for listing and creating tasks
	http.HandleFunc("/api/tasks", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Route based on HTTP method
		switch r.Method {
		case "GET":
			handlers.GetTasks(w, r)    // Get all tasks for user
		case "POST":
			handlers.CreateTask(w, r)  // Create new task
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method not allowed"}`))
		}
	}))
	
	// Handle /api/tasks/{id} (with trailing slash) - for individual task operations
	http.HandleFunc("/api/tasks/", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Route to appropriate handler based on HTTP method
		switch r.Method {
		case "GET":
			handlers.GetTask(w, r)     // Get specific task
		case "PUT":
			handlers.UpdateTask(w, r)  // Update specific task
		case "DELETE":
			handlers.DeleteTask(w, r)  // Delete specific task
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method not allowed"}`))
		}
	}))

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

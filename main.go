package main

import (
	"log"
	"net/http"

	"github.com/kcansari/task-management-api/config"
	"github.com/kcansari/task-management-api/database"
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

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

package database

import (
	"fmt"
	"log"

	"github.com/kcansari/task-management-api/models"
)

func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	log.Println("Running database migrations...")
	
	if err := DB.AutoMigrate(&models.User{}, &models.Task{}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	log.Println("Database migrations completed successfully")
	return nil
}

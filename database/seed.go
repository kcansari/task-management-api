package database

import (
	"fmt"
	"log"

	"github.com/kcansari/task-management-api/models"
)

func SeedData() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	log.Println("Seeding sample data...")

	sampleUser := models.User{
		Email:    "test@example.com",
		Password: "hashedpassword123",
	}

	if err := DB.Create(&sampleUser).Error; err != nil {
		return fmt.Errorf("failed to create sample user: %w", err)
	}

	sampleTasks := []models.Task{
		{
			Title:       "Complete project setup",
			Description: "Set up the basic project structure and database",
			Status:      models.TaskStatusCompleted,
			UserID:      sampleUser.ID,
		},
		{
			Title:       "Implement authentication",
			Description: "Add user registration and login functionality",
			Status:      models.TaskStatusInProgress,
			UserID:      sampleUser.ID,
		},
		{
			Title:       "Create API endpoints",
			Description: "Build REST API endpoints for task management",
			Status:      models.TaskStatusPending,
			UserID:      sampleUser.ID,
		},
	}

	for _, task := range sampleTasks {
		if err := DB.Create(&task).Error; err != nil {
			return fmt.Errorf("failed to create sample task: %w", err)
		}
	}

	log.Println("Sample data seeded successfully")
	return nil
}
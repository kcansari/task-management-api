package database

import (
	"log"

	"github.com/kcansari/task-management-api/config"
)

func Initialize(cfg *config.Config) error {
	if err := Connect(cfg); err != nil {
		return err
	}

	if err := RunMigrations(); err != nil {
		return err
	}

	if err := SeedData(); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

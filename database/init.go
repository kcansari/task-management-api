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

	log.Println("Database initialized successfully")
	return nil
}

package configs

import (
	"fmt"
	"log/slog"

	"jarvis/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDatabase creates a new GORM database connection and runs auto-migrations.
func NewDatabase(dsn string) (*gorm.DB, error) {
	slog.Info("connecting to database")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("running auto-migrations")
	if err := db.AutoMigrate(&models.Habit{}); err != nil {
		return nil, fmt.Errorf("failed to run auto-migration: %w", err)
	}

	return db, nil
}

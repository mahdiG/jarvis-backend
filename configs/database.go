package configs

import (
	"fmt"
	"log/slog"

	"jarvis/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDatabase creates a new GORM database connection from the application config
// and runs auto-migrations.
func NewDatabase() (*gorm.DB, error) {
	slog.Info("connecting to database")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		Envs.DatabaseHost,
		Envs.DatabasePort,
		Envs.DatabaseUser,
		Envs.DatabasePassword,
		Envs.DatabaseName,
		Envs.DatabaseSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("running auto-migrations")

	err = db.AutoMigrate(
		&models.Task{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to run auto-migration: %w", err)
	}

	return db, nil
}

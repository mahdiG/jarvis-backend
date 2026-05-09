package main

import (
	"log/slog"
	"os"

	"jarvis/configs"
	"jarvis/repositories"
	"jarvis/router"

	"github.com/gofiber/fiber/v2"
)

const defaultPort = "8080"

func main() {
	slog.Info("starting jarvis backend")

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=jarvis password=jarvis dbname=jarvis sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db, err := configs.NewDatabase(dsn)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	repositories.Init(db)

	app := fiber.New()

	router.Setup(app, db)

	slog.Info("server listening", "port", port)
	if err := app.Listen(":" + port); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

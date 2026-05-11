package main

import (
	"log/slog"
	"os"

	"jarvis/configs"
	"jarvis/repositories"
	"jarvis/router"
	"jarvis/services"
	"jarvis/utils"

	"github.com/gofiber/fiber/v2"
)

func main() {
	utils.InitSlog()

	slog.Info("starting jarvis backend")

	err := configs.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	db, err := configs.NewDatabase()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	repositories.Init(db)

	app := fiber.New()

	// Initialize the AI agent.
	agent, err := services.NewAgent()
	if err != nil {
		slog.Error("failed to initialize AI agent", "error", err)
		os.Exit(1)
	}

	router.Setup(app, db, agent)

	slog.Info("server listening", "port", configs.Envs.ServerPort)
	if err := app.Listen(":" + configs.Envs.ServerPort); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

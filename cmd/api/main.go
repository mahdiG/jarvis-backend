package main

import (
	"log/slog"
	"os"

	"jarvis/agent"
	"jarvis/configs"
	"jarvis/repositories"
	"jarvis/router"
	"jarvis/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	err := configs.LoadConfig()

	utils.InitSlog()

	slog.Info("starting jarvis backend")

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

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{configs.Envs.CORSAllowedOrigins},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Initialize the AI agent.
	err = agent.Init()
	if err != nil {
		slog.Error("failed to initialize AI agent", "error", err)
		os.Exit(1)
	}

	router.Setup(app)

	slog.Info("server listening", "port", configs.Envs.ServerPort)
	err = app.Listen(":" + configs.Envs.ServerPort)
	if err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

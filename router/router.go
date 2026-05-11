package router

import (
	"jarvis/controllers"
	"jarvis/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Setup configures all application routes.
func Setup(app *fiber.App, db *gorm.DB, agent *services.Agent) {

	v1 := app.Group("/v1")

	tasks := v1.Group("/tasks")
	tasks.Get("/", controllers.GetTasks)
	tasks.Post("/", controllers.CreateTask)
	tasks.Get("/:id", controllers.GetTask)
	tasks.Patch("/:id", controllers.UpdateTask)
	tasks.Delete("/:id", controllers.DeleteTask)

	chat := v1.Group("/chat")
	chat.Post("/", controllers.ChatHandler(agent))

}

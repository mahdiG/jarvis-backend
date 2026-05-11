package router

import (
	"jarvis/controllers"

	"github.com/gofiber/fiber/v2"
)

// Setup configures all application routes.
func Setup(app *fiber.App) {

	v1 := app.Group("/v1")

	tasks := v1.Group("/tasks")
	tasks.Get("/", controllers.GetTasks)
	tasks.Post("/", controllers.CreateTask)
	tasks.Get("/:id", controllers.GetTask)
	tasks.Patch("/:id", controllers.UpdateTask)
	tasks.Delete("/:id", controllers.DeleteTask)

	chat := v1.Group("/chat")
	chat.Post("/", controllers.GetChatMessage)

}

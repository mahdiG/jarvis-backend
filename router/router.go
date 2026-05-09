package router

import (
	"jarvis/controllers"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Setup configures all application routes.
func Setup(app *fiber.App, db *gorm.DB) {

	v1 := app.Group("/v1")
	habits := v1.Group("/habits")

	habits.Get("/:id", controllers.GetHabit)
}

package controllers

import (
	"errors"
	"log/slog"

	"jarvis/models"
	"jarvis/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetHabit(c *fiber.Ctx) error {
	id := c.Params("id")

	uid := models.UID(id)
	habit, err := repositories.GetHabit(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "habit not found",
			})
		}

		slog.Error("failed to get habit", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(habit)
}

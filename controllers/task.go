package controllers

import (
	"errors"
	"log/slog"

	"jarvis/models"
	"jarvis/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetTasks(c *fiber.Ctx) error {
	tasks, err := repositories.GetTasks()
	if err != nil {
		slog.Error("failed to get tasks", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tasks from db",
		})
	}

	return c.JSON(tasks)
}

func GetTask(c *fiber.Ctx) error {
	id := c.Params("id")

	uid := models.UID(id)
	task, err := repositories.GetTask(uid)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "task not found",
			})
		}

		slog.Error("failed to get task", "id", id, "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get task from db",
		})
	}

	return c.JSON(task)
}

func CreateTask(c *fiber.Ctx) error {
	var task models.Task

	if !ParseAndValidate(c, &task) {
		return nil
	}

	err := repositories.CreateTask(&task)
	if err != nil {
		slog.Error("failed to create task", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create task in db",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

func UpdateTask(c *fiber.Ctx) error {
	id := c.Params("id")

	var updates models.Task
	if !ParseAndValidate(c, &updates) {
		return nil
	}

	uid := models.UID(id)
	task, err := repositories.UpdateTask(uid, updates)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "task not found",
			})
		}

		slog.Error("failed to update task", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update task",
		})
	}

	return c.JSON(task)
}

func DeleteTask(c *fiber.Ctx) error {
	id := c.Params("id")

	uid := models.UID(id)
	err := repositories.DeleteTask(uid)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "task not found",
			})
		}

		slog.Error("failed to delete task", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete task",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
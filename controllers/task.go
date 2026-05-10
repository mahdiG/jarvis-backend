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

// CreateTaskRequest is the expected JSON body for creating a task.
type CreateTaskRequest struct {
	Title       string     `json:"title"       validate:"required"`
	Description string     `json:"description"`
	ParentID    models.UID `json:"parent_id"`
}

func CreateTask(c *fiber.Ctx) error {
	var req CreateTaskRequest

	if !ParseAndValidate(c, &req) {
		return nil
	}

	task := models.Task{
		Title:       req.Title,
		Description: req.Description,
		ParentID:    req.ParentID,
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

// UpdateTaskRequest is the expected JSON body for updating a task.
// Only provided fields will be updated.
type UpdateTaskRequest struct {
	Title       *string     `json:"title"       validate:"omitempty"`
	Description *string     `json:"description"`
	ParentID    *models.UID `json:"parent_id"`
}

func UpdateTask(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateTaskRequest
	if !ParseAndValidate(c, &req) {
		return nil
	}

	updates := make(map[string]any)
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no fields to update",
		})
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

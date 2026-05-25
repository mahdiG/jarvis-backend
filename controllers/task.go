package controllers

import (
	"errors"
	"log/slog"

	"jarvis/models"
	"jarvis/repositories"

	"github.com/gofiber/fiber/v3"
)

// GetTasks returns all tasks
// @Summary      List all tasks
// @Tags         Tasks
// @Produce      json
// @Success      200  {object}  Response[[]models.Task]
// @Failure      500  {object}  Response[any]
// @Router       /tasks [get]
func GetTasks(c fiber.Ctx) error {
	tasks, err := repositories.GetTasks(0, 0)
	if err != nil {
		slog.Error("failed to get tasks", "error", err)

		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get tasks from db")
	}

	return SuccessResponse(c, fiber.StatusOK, tasks, nil)
}

// GetTask returns a single task by its ID
// @Summary      Get a task
// @Tags         Tasks
// @Produce      json
// @Param        id   path      string  true  "Task ID"
// @Success      200  {object}  Response[models.Task]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks/{id} [get]
func GetTask(c fiber.Ctx) error {
	id := c.Params("id")

	task, err := repositories.GetTask(models.Task{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "task not found")
		}

		slog.Error("failed to get task", "id", id, "error", err)

		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get task from db")
	}

	return SuccessResponse(c, fiber.StatusOK, task, nil)
}

// CreateTasks creates multiple tasks
// @Summary      Create tasks
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        body  body      []models.Task  true  "Tasks to create"
// @Success      201  {object}  Response[[]models.Task]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks [post]
func CreateTasks(c fiber.Ctx) error {
	var tasks []models.Task

	if !Validate(c, &tasks) {
		return nil
	}

	created, err := repositories.CreateTasks(tasks)
	if err != nil {
		slog.Error("failed to create tasks", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to create tasks in db")
	}

	return SuccessResponse(c, fiber.StatusCreated, created, nil)
}

// UpdateTasks performs a batch partial update on tasks
// @Summary      Update tasks
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        body  body      []models.Task  true  "Tasks to update (each must include id)"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks [patch]
func UpdateTasks(c fiber.Ctx) error {
	var tasks []models.Task

	if !Validate(c, &tasks) {
		return nil
	}

	err := repositories.UpdateTasks(tasks)
	if err != nil {
		slog.Error("failed to update tasks", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to update tasks")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

// GetTrashTasks returns all soft-deleted tasks
// @Summary      List trashed tasks
// @Tags         Tasks
// @Produce      json
// @Success      200  {object}  Response[[]models.Task]
// @Failure      500  {object}  Response[any]
// @Router       /tasks/trash [get]
func GetTrashTasks(c fiber.Ctx) error {
	tasks, err := repositories.GetTrashTasks(0, 0)
	if err != nil {
		slog.Error("failed to get trashed tasks", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get trashed tasks from db")
	}

	return SuccessResponse(c, fiber.StatusOK, tasks, nil)
}

// RestoreTasks restores soft-deleted tasks by their IDs
// @Summary      Restore trashed tasks
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        body  body      []string  true  "Task IDs to restore"
// @Success      200  {object}  Response[[]models.Task]
// @Failure      400  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks/trash [patch]
func RestoreTasks(c fiber.Ctx) error {
	var ids []models.UID

	if !Validate(c, &ids) {
		return nil
	}

	tasks, err := repositories.RestoreTasks(ids)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "no matching tasks found in trash")
		}

		slog.Error("failed to restore tasks", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to restore tasks")
	}

	return SuccessResponse(c, fiber.StatusOK, tasks, nil)
}

// HardDeleteTasks permanently deletes tasks by their IDs
// @Summary      Permanently delete tasks
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        body  body      []string  true  "Task IDs to permanently delete"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks/trash [delete]
func HardDeleteTasks(c fiber.Ctx) error {
	var ids []models.UID

	if !Validate(c, &ids) {
		return nil
	}

	err := repositories.HardDeleteTasks(ids)
	if err != nil {
		slog.Error("failed to permanently delete tasks", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to permanently delete tasks")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

// SoftDeleteTasks soft-deletes tasks by their IDs
// @Summary      Delete tasks
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        body  body      []string  true  "Task IDs to delete"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks [delete]
func SoftDeleteTasks(c fiber.Ctx) error {
	var ids []models.UID

	if !Validate(c, &ids) {
		return nil
	}

	err := repositories.SoftDeleteTasks(ids)
	if err != nil {
		slog.Error("failed to delete tasks", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete tasks")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}
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

// CreateTask creates a new task
// @Summary      Create a task
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        body  body      models.Task  true  "Task to create"
// @Success      201  {object}  Response[models.Task]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks [post]
func CreateTask(c fiber.Ctx) error {
	var task models.Task

	if !Validate(c, &task) {
		return nil
	}

	task, err := repositories.CreateTask(task)
	if err != nil {
		slog.Error("failed to create task", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to create task in db")
	}

	return SuccessResponse(c, fiber.StatusCreated, task, nil)
}

// UpdateTask updates an existing task (partial update)
// @Summary      Update a task
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        id    path      string       true  "Task ID"
// @Param        body  body      models.Task  true  "Updated task fields"
// @Success      200  {object}  Response[models.Task]
// @Failure      400  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks/{id} [patch]
func UpdateTask(c fiber.Ctx) error {
	id := c.Params("id")

	var input models.Task
	if !Validate(c, &input) {
		return nil
	}

	input.ID = models.UID(id)

	task, err := repositories.UpdateTask(input)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "task not found")
		}

		slog.Error("failed to update task", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to update task")
	}

	return SuccessResponse(c, fiber.StatusOK, task, nil)
}

// DeleteTask deletes a task by its ID
// @Summary      Delete a task
// @Tags         Tasks
// @Produce      json
// @Param        id   path      string  true  "Task ID"
// @Success      200  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tasks/{id} [delete]
func DeleteTask(c fiber.Ctx) error {
	id := c.Params("id")

	err := repositories.DeleteTask(models.Task{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "task not found")
		}

		slog.Error("failed to delete task", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete task")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

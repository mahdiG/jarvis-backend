package controllers

import (
	"errors"
	"log/slog"

	"jarvis/models"
	"jarvis/repositories"

	"github.com/gofiber/fiber/v3"
)

// GetNotes returns all notes
// @Summary      List all notes
// @Tags         Notes
// @Produce      json
// @Success      200  {array}   models.Note
// @Failure      500  {object}  fiber.Map
// @Router       /notes [get]
func GetNotes(c fiber.Ctx) error {
	notes, err := repositories.GetNotes(0, 0)
	if err != nil {
		slog.Error("failed to get notes", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get notes from db",
		})
	}

	return c.JSON(notes)
}

// GetNote returns a single note by its ID
// @Summary      Get a note
// @Tags         Notes
// @Produce      json
// @Param        id   path      string  true  "Note ID"
// @Success      200  {object}  models.Note
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /notes/{id} [get]
func GetNote(c fiber.Ctx) error {
	id := c.Params("id")

	note, err := repositories.GetNote(models.Note{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "note not found",
			})
		}

		slog.Error("failed to get note", "id", id, "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get note from db",
		})
	}

	return c.JSON(note)
}

// CreateNote creates a new note
// @Summary      Create a note
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        body  body      models.Note  true  "Note to create"
// @Success      201  {object}  models.Note
// @Failure      400  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /notes [post]
func CreateNote(c fiber.Ctx) error {
	var note models.Note

	if !Validate(c, &note) {
		return nil
	}

	note, err := repositories.CreateNote(note)
	if err != nil {
		slog.Error("failed to create note", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create note in db",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(note)
}

// UpdateNote updates an existing note (partial update)
// @Summary      Update a note
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        id    path      string       true  "Note ID"
// @Param        body  body      models.Note  true  "Updated note fields"
// @Success      200  {object}  models.Note
// @Failure      400  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /notes/{id} [patch]
func UpdateNote(c fiber.Ctx) error {
	id := c.Params("id")

	var input models.Note
	if !Validate(c, &input) {
		return nil
	}

	input.ID = models.UID(id)

	note, err := repositories.UpdateNote(input)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "note not found",
			})
		}

		slog.Error("failed to update note", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update note",
		})
	}

	return c.JSON(note)
}

// DeleteNote deletes a note by its ID
// @Summary      Delete a note
// @Tags         Notes
// @Produce      json
// @Param        id   path      string  true  "Note ID"
// @Success      204  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /notes/{id} [delete]
func DeleteNote(c fiber.Ctx) error {
	id := c.Params("id")

	err := repositories.DeleteNote(models.Note{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "note not found",
			})
		}

		slog.Error("failed to delete note", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete note",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
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
// @Success      200  {object}  Response[[]models.Note]
// @Failure      500  {object}  Response[any]
// @Router       /notes [get]
func GetNotes(c fiber.Ctx) error {
	notes, err := repositories.GetNotes(0, 0)
	if err != nil {
		slog.Error("failed to get notes", "error", err)

		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get notes from db")
	}

	return SuccessResponse(c, fiber.StatusOK, notes, nil)
}

// GetNote returns a single note by its ID
// @Summary      Get a note
// @Tags         Notes
// @Produce      json
// @Param        id   path      string  true  "Note ID"
// @Success      200  {object}  Response[models.Note]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes/{id} [get]
func GetNote(c fiber.Ctx) error {
	id := c.Params("id")

	note, err := repositories.GetNote(models.Note{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "note not found")
		}

		slog.Error("failed to get note", "id", id, "error", err)

		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get note from db")
	}

	return SuccessResponse(c, fiber.StatusOK, note, nil)
}

// CreateNote creates a new note
// @Summary      Create a note
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        body  body      models.Note  true  "Note to create"
// @Success      201  {object}  Response[models.Note]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes [post]
func CreateNote(c fiber.Ctx) error {
	var note models.Note

	if !Validate(c, &note) {
		return nil
	}

	note, err := repositories.CreateNote(note)
	if err != nil {
		slog.Error("failed to create note", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to create note in db")
	}

	return SuccessResponse(c, fiber.StatusCreated, note, nil)
}

// UpdateNote updates an existing note (partial update)
// @Summary      Update a note
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        id    path      string       true  "Note ID"
// @Param        body  body      models.Note  true  "Updated note fields"
// @Success      200  {object}  Response[models.Note]
// @Failure      400  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
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
			return ErrorResponse(c, fiber.StatusNotFound, "note not found")
		}

		slog.Error("failed to update note", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to update note")
	}

	return SuccessResponse(c, fiber.StatusOK, note, nil)
}

// GetTrashNotes returns all soft-deleted notes
// @Summary      List trashed notes
// @Tags         Notes
// @Produce      json
// @Success      200  {object}  Response[[]models.Note]
// @Failure      500  {object}  Response[any]
// @Router       /notes/trash [get]
func GetTrashNotes(c fiber.Ctx) error {
	notes, err := repositories.GetTrashNotes(0, 0)
	if err != nil {
		slog.Error("failed to get trashed notes", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get trashed notes from db")
	}

	return SuccessResponse(c, fiber.StatusOK, notes, nil)
}

// RestoreNote restores a soft-deleted note by its ID
// @Summary      Restore a trashed note
// @Tags         Notes
// @Produce      json
// @Param        id   path      string  true  "Note ID"
// @Success      200  {object}  Response[models.Note]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes/trash/{id} [patch]
func RestoreNote(c fiber.Ctx) error {
	id := c.Params("id")

	note, err := repositories.RestoreNote(models.UID(id))
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "note not found in trash")
		}

		slog.Error("failed to restore note", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to restore note")
	}

	return SuccessResponse(c, fiber.StatusOK, note, nil)
}

// PermanentDeleteNote permanently deletes a note by its ID (even if soft-deleted)
// @Summary      Permanently delete a note
// @Tags         Notes
// @Produce      json
// @Param        id   path      string  true  "Note ID"
// @Success      200  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes/trash/{id} [delete]
func PermanentDeleteNote(c fiber.Ctx) error {
	id := c.Params("id")

	err := repositories.PermanentDeleteNote(models.UID(id))
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "note not found")
		}

		slog.Error("failed to permanently delete note", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to permanently delete note")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

// DeleteNote deletes a note by its ID
// @Summary      Delete a note
// @Tags         Notes
// @Produce      json
// @Param        id   path      string  true  "Note ID"
// @Success      200  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes/{id} [delete]
func DeleteNote(c fiber.Ctx) error {
	id := c.Params("id")

	err := repositories.DeleteNote(models.Note{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "note not found")
		}

		slog.Error("failed to delete note", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete note")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

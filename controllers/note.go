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

// CreateNotes creates multiple notes
// @Summary      Create notes
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        body  body      []models.Note  true  "Notes to create"
// @Success      201  {object}  Response[[]models.Note]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes [post]
func CreateNotes(c fiber.Ctx) error {
	var notes []models.Note

	if !Validate(c, &notes) {
		return nil
	}

	created, err := repositories.CreateNotes(notes)
	if err != nil {
		slog.Error("failed to create notes", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to create notes in db")
	}

	return SuccessResponse(c, fiber.StatusCreated, created, nil)
}

// UpdateNotes performs a batch partial update on notes
// @Summary      Update notes
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        body  body      []models.Note  true  "Notes to update (each must include id)"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes [patch]
func UpdateNotes(c fiber.Ctx) error {
	var notes []models.Note

	if !Validate(c, &notes) {
		return nil
	}

	err := repositories.UpdateNotes(notes)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "one or more notes not found")
		}

		slog.Error("failed to update notes", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to update notes")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
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

// RestoreNotes restores soft-deleted notes by their IDs
// @Summary      Restore trashed notes
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        body  body      []string  true  "Note IDs to restore"
// @Success      200  {object}  Response[[]models.Note]
// @Failure      400  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes/trash [patch]
func RestoreNotes(c fiber.Ctx) error {
	var ids []models.UID

	if !Validate(c, &ids) {
		return nil
	}

	notes, err := repositories.RestoreNotes(ids)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "no matching notes found in trash")
		}

		slog.Error("failed to restore notes", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to restore notes")
	}

	return SuccessResponse(c, fiber.StatusOK, notes, nil)
}

// HardDeleteNotes permanently deletes notes by their IDs
// @Summary      Permanently delete notes
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        body  body      []string  true  "Note IDs to permanently delete"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes/trash [delete]
func HardDeleteNotes(c fiber.Ctx) error {
	var ids []models.UID

	if !Validate(c, &ids) {
		return nil
	}

	err := repositories.HardDeleteNotes(ids)
	if err != nil {
		slog.Error("failed to permanently delete notes", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to permanently delete notes")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

// SoftDeleteNotes soft-deletes notes by their IDs
// @Summary      Delete notes
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        body  body      []string  true  "Note IDs to delete"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /notes [delete]
func SoftDeleteNotes(c fiber.Ctx) error {
	var ids []models.UID

	if !Validate(c, &ids) {
		return nil
	}

	err := repositories.SoftDeleteNotes(ids)
	if err != nil {
		slog.Error("failed to delete notes", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete notes")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}
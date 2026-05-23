package controllers

import (
	"errors"
	"log/slog"

	"jarvis/models"
	"jarvis/repositories"

	"github.com/gofiber/fiber/v3"
)

// GetTags returns all tags
// @Summary      List all tags
// @Tags         Tags
// @Produce      json
// @Success      200  {object}  Response[[]models.Tag]
// @Failure      500  {object}  Response[any]
// @Router       /tags [get]
func GetTags(c fiber.Ctx) error {
	tags, err := repositories.GetTags(0, 0)
	if err != nil {
		slog.Error("failed to get tags", "error", err)

		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get tags from db")
	}

	return SuccessResponse(c, fiber.StatusOK, tags, nil)
}

// GetTag returns a single tag by its ID
// @Summary      Get a tag
// @Tags         Tags
// @Produce      json
// @Param        id   path      string  true  "Tag ID"
// @Success      200  {object}  Response[models.Tag]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tags/{id} [get]
func GetTag(c fiber.Ctx) error {
	id := c.Params("id")

	tag, err := repositories.GetTag(models.Tag{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "tag not found")
		}

		slog.Error("failed to get tag", "id", id, "error", err)

		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to get tag from db")
	}

	return SuccessResponse(c, fiber.StatusOK, tag, nil)
}

// CreateTags creates multiple tags
// @Summary      Create tags
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        body  body      []models.Tag  true  "Tags to create"
// @Success      201  {object}  Response[[]models.Tag]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tags [post]
func CreateTags(c fiber.Ctx) error {
	var tags []models.Tag

	if !Validate(c, &tags) {
		return nil
	}

	created, err := repositories.CreateTags(tags)
	if err != nil {
		slog.Error("failed to create tags", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to create tags in db")
	}

	return SuccessResponse(c, fiber.StatusCreated, created, nil)
}

// UpdateTags performs a batch partial update on tags
// @Summary      Update tags
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        body  body      []models.Tag  true  "Tags to update (each must include id)"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tags [patch]
func UpdateTags(c fiber.Ctx) error {
	var tags []models.Tag

	if !Validate(c, &tags) {
		return nil
	}

	err := repositories.UpdateTags(tags)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "one or more tags not found")
		}

		slog.Error("failed to update tags", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to update tags")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

// DeleteTags deletes tags by their IDs
// @Summary      Delete tags
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        body  body      []string  true  "Tag IDs to delete"
// @Success      200  {object}  Response[any]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tags [delete]
func DeleteTags(c fiber.Ctx) error {
	var ids []models.UID

	if !Validate(c, &ids) {
		return nil
	}

	err := repositories.DeleteTags(ids)
	if err != nil {
		slog.Error("failed to delete tags", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete tags")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}
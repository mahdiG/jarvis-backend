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

// CreateTag creates a new tag
// @Summary      Create a tag
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        body  body      models.Tag  true  "Tag to create"
// @Success      201  {object}  Response[models.Tag]
// @Failure      400  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tags [post]
func CreateTag(c fiber.Ctx) error {
	var tag models.Tag

	if !Validate(c, &tag) {
		return nil
	}

	tag, err := repositories.CreateTag(tag)
	if err != nil {
		slog.Error("failed to create tag", "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to create tag in db")
	}

	return SuccessResponse(c, fiber.StatusCreated, tag, nil)
}

// UpdateTag updates an existing tag (partial update)
// @Summary      Update a tag
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        id    path      string       true  "Tag ID"
// @Param        body  body      models.Tag  true  "Updated tag fields"
// @Success      200  {object}  Response[models.Tag]
// @Failure      400  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tags/{id} [patch]
func UpdateTag(c fiber.Ctx) error {
	id := c.Params("id")

	var input models.Tag
	if !Validate(c, &input) {
		return nil
	}

	input.ID = models.UID(id)

	tag, err := repositories.UpdateTag(input)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "tag not found")
		}

		slog.Error("failed to update tag", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to update tag")
	}

	return SuccessResponse(c, fiber.StatusOK, tag, nil)
}

// DeleteTag deletes a tag by its ID
// @Summary      Delete a tag
// @Tags         Tags
// @Produce      json
// @Param        id   path      string  true  "Tag ID"
// @Success      200  {object}  Response[any]
// @Failure      404  {object}  Response[any]
// @Failure      500  {object}  Response[any]
// @Router       /tags/{id} [delete]
func DeleteTag(c fiber.Ctx) error {
	id := c.Params("id")

	err := repositories.DeleteTag(models.Tag{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "tag not found")
		}

		slog.Error("failed to delete tag", "id", id, "error", err)
		return ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete tag")
	}

	return SuccessResponse[any](c, fiber.StatusOK, nil, nil)
}

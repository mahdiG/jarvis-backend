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
// @Success      200  {array}   models.Tag
// @Failure      500  {object}  fiber.Map
// @Router       /tags [get]
func GetTags(c fiber.Ctx) error {
	tags, err := repositories.GetTags(0, 0)
	if err != nil {
		slog.Error("failed to get tags", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tags from db",
		})
	}

	return c.JSON(tags)
}

// GetTag returns a single tag by its ID
// @Summary      Get a tag
// @Tags         Tags
// @Produce      json
// @Param        id   path      string  true  "Tag ID"
// @Success      200  {object}  models.Tag
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /tags/{id} [get]
func GetTag(c fiber.Ctx) error {
	id := c.Params("id")

	tag, err := repositories.GetTag(models.Tag{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tag not found",
			})
		}

		slog.Error("failed to get tag", "id", id, "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tag from db",
		})
	}

	return c.JSON(tag)
}

// CreateTag creates a new tag
// @Summary      Create a tag
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        body  body      models.Tag  true  "Tag to create"
// @Success      201  {object}  models.Tag
// @Failure      400  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /tags [post]
func CreateTag(c fiber.Ctx) error {
	var tag models.Tag

	if !Validate(c, &tag) {
		return nil
	}

	tag, err := repositories.CreateTag(tag)
	if err != nil {
		slog.Error("failed to create tag", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create tag in db",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tag)
}

// UpdateTag updates an existing tag (partial update)
// @Summary      Update a tag
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        id    path      string       true  "Tag ID"
// @Param        body  body      models.Tag  true  "Updated tag fields"
// @Success      200  {object}  models.Tag
// @Failure      400  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
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
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tag not found",
			})
		}

		slog.Error("failed to update tag", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update tag",
		})
	}

	return c.JSON(tag)
}

// DeleteTag deletes a tag by its ID
// @Summary      Delete a tag
// @Tags         Tags
// @Produce      json
// @Param        id   path      string  true  "Tag ID"
// @Success      204  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /tags/{id} [delete]
func DeleteTag(c fiber.Ctx) error {
	id := c.Params("id")

	err := repositories.DeleteTag(models.Tag{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tag not found",
			})
		}

		slog.Error("failed to delete tag", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete tag",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
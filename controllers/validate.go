package controllers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// ParseAndValidate parses the request body into the target struct and validates it
// using the go-playground/validator library. If parsing or validation fails, it sends
// a JSON error response and returns false. On success it returns true.
//
// Usage:
//
//	var req CreateTaskRequest
//	if !ParseAndValidate(c, &req) {
//	    return nil
//	}
func ParseAndValidate[T any](c *fiber.Ctx, target *T) bool {
	err := c.BodyParser(target)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
		return false
	}

	err = validate.Struct(target)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "request validation failed",
			})
			return false
		}

		fields := make([]fiber.Map, 0, len(validationErrors))
		for _, fieldError := range validationErrors {
			fields = append(fields, fiber.Map{
				"field":   fieldError.Field(),
				"tag":     fieldError.Tag(),
				"message": fieldError.Error(),
			})
		}

		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "request validation failed",
			"fields": fields,
		})
		return false
	}

	return true
}
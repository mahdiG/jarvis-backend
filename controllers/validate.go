package controllers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var validate = validator.New()

// Validate parses the request body into the target struct and validates it
// using the go-playground/validator library. If parsing or validation fails, it sends
// a JSON error response and returns false. On success it returns true.
//
// Usage:
//
//	var req CreateTaskRequest
//	if !Validate(c, &req) {
//	    return nil
//	}
func Validate[Type any](c fiber.Ctx, target *Type) bool {
	err := c.Bind().Body(target)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(Response[any]{
			Data: nil,
			Error: &ResponseError{
				Message: "invalid request body",
			},
		})
		return false
	}

	err = validate.Struct(target)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			c.Status(fiber.StatusBadRequest).JSON(Response[any]{
				Data: nil,
				Error: &ResponseError{
					Message: "request validation failed",
				},
			})
			return false
		}

		fields := make([]ResponseErrorField, 0, len(validationErrors))
		for _, fieldError := range validationErrors {
			fields = append(fields, ResponseErrorField{
				Field:   fieldError.Field(),
				Tag:     fieldError.Tag(),
				Message: fieldError.Error(),
			})
		}

		c.Status(fiber.StatusBadRequest).JSON(Response[any]{
			Data: nil,
			Error: &ResponseError{
				Message: "request validation failed",
				Fields:  fields,
			},
		})
		return false
	}

	return true
}

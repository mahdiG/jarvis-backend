package controllers

import (
	"reflect"

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
// Validate parses the request body into target and validates it.
// Type may be a struct or a slice of structs. For slices each element is
// validated individually. Returns false and sends an error response on failure.
func Validate[Type any](c fiber.Ctx, target *Type) bool {
	err := c.Bind().Body(target)
	if err != nil {
		ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
		return false
	}

	return validateValue(c, target)
}

// validateValue performs validation on a struct or a slice of structs.
func validateValue(c fiber.Ctx, value any) bool {
	reflectedValue := reflect.ValueOf(value)
	if reflectedValue.Kind() == reflect.Ptr && reflectedValue.Elem().Kind() == reflect.Slice {
		sliceReflect := reflectedValue.Elem()
		for index := 0; index < sliceReflect.Len(); index++ {
			element := sliceReflect.Index(index)
			if element.Kind() == reflect.Ptr {
				element = element.Elem()
			}
			// Only validate struct elements; primitive types (e.g. UID) need no struct validation.
			if element.Kind() == reflect.Struct {
				if !validateStruct(c, element.Addr().Interface()) {
					return false
				}
			}
		}
		return true
	}
	return validateStruct(c, value)
}

// validateStruct validates a single struct and sends an error response on failure.
func validateStruct(c fiber.Ctx, target any) bool {
	err := validate.Struct(target)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			ErrorResponse(c, fiber.StatusBadRequest, "request validation failed")
			return false
		}

		fields := make([]ResponseErrorField, 0, len(validationErrors))
		for _, fieldError := range validationErrors {
			fields = append(fields, ResponseErrorField{
				Name:    fieldError.Field(),
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

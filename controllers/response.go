package controllers

import "github.com/gofiber/fiber/v3"

// ─── Core Response Envelope ────────────────────────────────────────────

// Response is the standard API response envelope.
// Type is the concrete type of the successful payload.
type Response[DataType any] struct {
	Data  DataType
	Error *ResponseError
	Meta  *ResponseMeta
}

// ─── Error Structures ──────────────────────────────────────────────────

// ResponseError represents an API error.
type ResponseError struct {
	// Message is a human‑readable description of the error.
	Message string

	// Fields contains per‑field validation errors (usually for 422 responses).
	Fields []ResponseErrorField
}

// ResponseErrorField describes a single field validation error.
type ResponseErrorField struct {
	Field   string // the JSON fname
	Tag     string // validation tag that failed
	Message string // human‑readable reason
}

// ─── Metadata ──────────────────────────────────────────────────────────

// ResponseMeta holds additional response metadata such as pagination and request tracing.
type ResponseMeta struct {
	// RequestID can be used for tracing requests across services.
	RequestID string

	// Pagination contains cursor/offset‑based pagination info.
	// It is nil for endpoints that do not paginate.
	Pagination *Pagination

	// Warnings contains non‑fatal messages (e.g. “this endpoint is deprecated”).
	Warnings []string
}

// Pagination provides pagination metadata.
type Pagination struct {
	// Page is the current page number (1‑based).
	Page int

	// PageSize is the maximum number of items per page.
	PageSize int

	// TotalItems is the total number of items across all pages.
	TotalItems int64

	// TotalPages is the total number of pages (TotalItems / PageSize).
	TotalPages int
}

// SuccessResponse builds a success response (no error, no meta) with data.
func SuccessResponse[dataType any](c fiber.Ctx, statusCode int, data dataType, metadata *ResponseMeta) error {
	return c.Status(statusCode).JSON(Response[dataType]{
		Data: data,
		Meta: metadata,
	})
}

// ErrorResponse builds an error response with a message and optional code.
func ErrorResponse(c fiber.Ctx, statusCode int, errorMessage string) error {
	return c.Status(statusCode).JSON(Response[any]{
		Data: nil,
		Error: &ResponseError{
			Message: errorMessage,
		},
	})
}

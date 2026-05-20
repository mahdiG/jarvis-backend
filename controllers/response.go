package controllers

import "github.com/gofiber/fiber/v3"

// ─── Core Response Envelope ────────────────────────────────────────────

// Response is the standard API response envelope.
// Type is the concrete type of the successful payload.
type Response[DataType any] struct {
	Data  DataType       `json:"data"`
	Error *ResponseError `json:"error,omitempty"`
	Meta  *ResponseMeta  `json:"meta,omitempty"`
}

// ─── Error Structures ──────────────────────────────────────────────────

// ResponseError represents an API error.
type ResponseError struct {
	// Message is a human‑readable description of the error.
	Message string `json:"message"`

	// Fields contains per‑field validation errors (usually for 422 responses).
	Fields []ResponseErrorField `json:"fields,omitempty"`
}

// ResponseErrorField describes a single field validation error.
type ResponseErrorField struct {
	Name    string `json:"name"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ─── Metadata ──────────────────────────────────────────────────────────

// ResponseMeta holds additional response metadata such as pagination and request tracing.
type ResponseMeta struct {
	// RequestID can be used for tracing requests across services.
	RequestID string `json:"request_id,omitempty"`

	// Pagination contains cursor/offset‑based pagination info.
	// It is nil for endpoints that do not paginate.
	Pagination *Pagination `json:"pagination,omitempty"`

	// Warnings contains non‑fatal messages (e.g. “this endpoint is deprecated”).
	Warnings []string `json:"warnings,omitempty"`
}

// Pagination provides pagination metadata.
type Pagination struct {
	// Page is the current page number (1‑based).
	Page int `json:"page"`

	// PageSize is the maximum number of items per page.
	PageSize int `json:"page_size"`

	// TotalItems is the total number of items across all pages.
	TotalItems int64 `json:"total_items"`

	// TotalPages is the total number of pages (TotalItems / PageSize).
	TotalPages int `json:"total_pages"`
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

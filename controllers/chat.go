package controllers

import (
	"log/slog"

	"jarvis/agent"

	"github.com/gofiber/fiber/v2"
)

// ChatRequest is the expected JSON body for the chat endpoint.
type ChatRequest struct {
	// Messages is the conversation history so far.
	// The last message should be from the user.
	Messages []agent.Message `validate:"required,min=1"`
}

// SendChatMessage handles the chat endpoint.
func SendChatMessage(c *fiber.Ctx) error {
	var request ChatRequest

	if !Validate(c, &request) {
		return nil
	}

	response, err := agent.Chat(c.Context(), request.Messages)
	if err != nil {
		slog.Error("failed to process chat message", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process chat message",
		})
	}

	return c.JSON(response)
}

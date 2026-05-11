package controllers

import (
	"jarvis/services"

	"github.com/gofiber/fiber/v2"
)

// ChatRequest is the expected JSON body for the chat endpoint.
type ChatRequest struct {
	// Messages is the conversation history so far.
	// The last message should be from the user.
	Messages []services.AgentMessage `json:"messages" validate:"required,min=1"`
}

// ChatHandler creates a handler for the chat endpoint using the given agent.
func ChatHandler(agent *services.Agent) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ChatRequest
		if !Validate(c, &req) {
			return nil
		}

		resp, err := agent.Chat(c.Context(), req.Messages)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to process chat message: " + err.Error(),
			})
		}

		return c.JSON(resp)
	}
}
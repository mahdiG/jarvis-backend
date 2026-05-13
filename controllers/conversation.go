package controllers

import (
	"errors"
	"log/slog"

	"jarvis/agent"
	"jarvis/models"
	"jarvis/repositories"

	"github.com/gofiber/fiber/v3"
)

// --------------------------------------------------------------------------
// Conversation CRUD
// --------------------------------------------------------------------------

// GetConversations returns all conversations
// @Summary      List all conversations
// @Tags         Conversations
// @Produce      json
// @Success      200  {array}   models.Conversation
// @Failure      500  {object}  fiber.Map
// @Router       /conversations [get]
func GetConversations(c fiber.Ctx) error {
	conversations, err := repositories.GetConversations(0, 0)
	if err != nil {
		slog.Error("failed to get conversations", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get conversations from db",
		})
	}

	return c.JSON(conversations)
}

// GetConversation returns a single conversation by its ID
// @Summary      Get a conversation
// @Tags         Conversations
// @Produce      json
// @Param        id   path      string  true  "Conversation ID"
// @Success      200  {object}  models.Conversation
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /conversations/{id} [get]
func GetConversation(c fiber.Ctx) error {
	id := c.Params("id")

	conversation, err := repositories.GetConversation(models.Conversation{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "conversation not found",
			})
		}

		slog.Error("failed to get conversation", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get conversation from db",
		})
	}

	return c.JSON(conversation)
}

// CreateConversation creates a new conversation
// @Summary      Create a conversation
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Param        body  body      models.Conversation  true  "Conversation to create"
// @Success      201  {object}  models.Conversation
// @Failure      400  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /conversations [post]
func CreateConversation(c fiber.Ctx) error {
	var conversation models.Conversation

	if !Validate(c, &conversation) {
		return nil
	}

	conversation, err := repositories.CreateConversation(conversation)
	if err != nil {
		slog.Error("failed to create conversation", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create conversation in db",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(conversation)
}

// UpdateConversation updates an existing conversation (partial update)
// @Summary      Update a conversation
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "Conversation ID"
// @Param        body  body      models.Conversation  true  "Updated conversation fields"
// @Success      200  {object}  models.Conversation
// @Failure      400  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /conversations/{id} [patch]
func UpdateConversation(c fiber.Ctx) error {
	id := c.Params("id")

	var input models.Conversation
	if !Validate(c, &input) {
		return nil
	}

	input.ID = models.UID(id)

	conversation, err := repositories.UpdateConversation(input)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "conversation not found",
			})
		}

		slog.Error("failed to update conversation", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update conversation",
		})
	}

	return c.JSON(conversation)
}

// DeleteConversation deletes a conversation by its ID
// @Summary      Delete a conversation
// @Tags         Conversations
// @Produce      json
// @Param        id   path      string  true  "Conversation ID"
// @Success      204  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /conversations/{id} [delete]
func DeleteConversation(c fiber.Ctx) error {
	id := c.Params("id")

	err := repositories.DeleteConversation(models.Conversation{Base: models.Base{ID: models.UID(id)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "conversation not found",
			})
		}

		slog.Error("failed to delete conversation", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete conversation",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// --------------------------------------------------------------------------
// Messages
// --------------------------------------------------------------------------

// GetMessages returns all messages for a conversation, ordered by creation time.
// @Summary      List messages in a conversation
// @Tags         Messages
// @Produce      json
// @Param        id  path      string  true  "Conversation ID"
// @Success      200  {array}  models.Message
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /conversations/{id}/messages [get]
func GetMessages(c fiber.Ctx) error {
	conversationID := c.Params("id")

	// Verify the conversation exists.
	_, err := repositories.GetConversation(models.Conversation{Base: models.Base{ID: models.UID(conversationID)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "conversation not found",
			})
		}
		slog.Error("failed to get conversation for messages", "id", conversationID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get messages",
		})
	}

	messages, err := repositories.GetMessages(models.UID(conversationID))
	if err != nil {
		slog.Error("failed to get messages", "conversation_id", conversationID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get messages from db",
		})
	}

	return c.JSON(messages)
}

// SendMessage handles posting a user message to a conversation.
// It saves the user message, calls the AI agent with the full conversation
// history, saves the assistant response, and returns both messages.
// @Summary      Send a message to a conversation
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        id    path      string         true  "Conversation ID"
// @Param        body  body      models.Message  true  "User message (content field)"
// @Success      201  {object}  fiber.Map  "Returns { user: Message, assistant: Message }"
// @Failure      400  {object}  fiber.Map
// @Failure      404  {object}  fiber.Map
// @Failure      500  {object}  fiber.Map
// @Router       /conversations/{id}/messages [post]
func SendMessage(c fiber.Ctx) error {
	conversationID := c.Params("id")

	// Verify the conversation exists.
	_, err := repositories.GetConversation(models.Conversation{Base: models.Base{ID: models.UID(conversationID)}})
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "conversation not found",
			})
		}
		slog.Error("failed to get conversation for send", "id", conversationID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process message",
		})
	}

	// Parse the incoming message.
	var input models.Message
	if !Validate(c, &input) {
		return nil
	}

	input.ConversationID = models.UID(conversationID)
	input.Role = "user"

	// Save the user message.
	userMessage, err := repositories.CreateMessage(input)
	if err != nil {
		slog.Error("failed to save user message", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save message",
		})
	}

	// Load all messages in the conversation to build context.
	allMessages, err := repositories.GetMessages(models.UID(conversationID))
	if err != nil {
		slog.Error("failed to load conversation messages", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process message",
		})
	}

	// Convert to agent messages.
	agentMessages := make([]agent.Message, len(allMessages))
	for i, m := range allMessages {
		agentMessages[i] = agent.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	// Call the AI agent.
	response, err := agent.Chat(c.Context(), agentMessages)
	if err != nil {
		slog.Error("failed to process chat message", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process chat message",
		})
	}

	// Save the assistant's response.
	assistantMessage, err := repositories.CreateMessage(models.Message{
		ConversationID: models.UID(conversationID),
		Role:           "assistant",
		Content:        response.Reply,
	})
	if err != nil {
		slog.Error("failed to save assistant message", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save response",
		})
	}

	// Return a response containing both messages.
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":      userMessage,
		"assistant": assistantMessage,
	})
}
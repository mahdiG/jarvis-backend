package controllers

import (
	"errors"
	"log/slog"

	"jarvis/agent"
	"jarvis/models"
	"jarvis/repositories"

	"github.com/gofiber/fiber/v2"
)

// --------------------------------------------------------------------------
// Conversation CRUD
// --------------------------------------------------------------------------

func GetConversations(c *fiber.Ctx) error {
	conversations, err := repositories.GetConversations(0, 0)
	if err != nil {
		slog.Error("failed to get conversations", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get conversations from db",
		})
	}

	return c.JSON(conversations)
}

func GetConversation(c *fiber.Ctx) error {
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

func CreateConversation(c *fiber.Ctx) error {
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

func UpdateConversation(c *fiber.Ctx) error {
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

func DeleteConversation(c *fiber.Ctx) error {
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
func GetMessages(c *fiber.Ctx) error {
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
func SendMessage(c *fiber.Ctx) error {
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
package controllers_test

import (
	"net/http"
	"testing"

	"jarvis/models"
	"jarvis/repositories"
)

// --------------------------------------------------------------------------
// TestGetMessages
// --------------------------------------------------------------------------

func TestGetMessages_Empty(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	conversation, err := repositories.CreateConversation(models.Conversation{Title: "Empty chat"})
	if err != nil {
		t.Fatalf("failed to seed conversation: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/conversations/"+string(conversation.ID)+"/messages", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var messages []models.Message
	err = DecodeJSON(resp, &messages)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected empty messages, got %d", len(messages))
	}
}

func TestGetMessages_WithItems(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	conversation, err := repositories.CreateConversation(models.Conversation{Title: "Chat with messages"})
	if err != nil {
		t.Fatalf("failed to seed conversation: %v", err)
	}

	// Seed two messages.
	_, err = repositories.CreateMessage(models.Message{
		ConversationID: conversation.ID,
		Role:           "user",
		Content:        "Hello",
	})
	if err != nil {
		t.Fatalf("failed to seed message: %v", err)
	}
	_, err = repositories.CreateMessage(models.Message{
		ConversationID: conversation.ID,
		Role:           "assistant",
		Content:        "Hi there!",
	})
	if err != nil {
		t.Fatalf("failed to seed message: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/conversations/"+string(conversation.ID)+"/messages", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var messages []models.Message
	err = DecodeJSON(resp, &messages)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Content != "Hello" {
		t.Errorf("expected first message content %q, got %q", "Hello", messages[0].Content)
	}
	if messages[1].Content != "Hi there!" {
		t.Errorf("expected second message content %q, got %q", "Hi there!", messages[1].Content)
	}
}

func TestGetMessages_ConversationNotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/conversations/nonexistent1/messages", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	var errResp ResponseError
	err = DecodeJSON(resp, &errResp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response")
	}
}
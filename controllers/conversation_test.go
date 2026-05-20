package controllers_test

import (
	"net/http"
	"testing"

	"jarvis/models"
	"jarvis/repositories"
)

// --------------------------------------------------------------------------
// TestCreateConversation
// --------------------------------------------------------------------------

func TestCreateConversation_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/conversations", `{"title":"Chat about coding"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var conversation models.Conversation
	_, err = DecodeResponseData(resp, &conversation)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if conversation.Title != "Chat about coding" {
		t.Errorf("expected title %q, got %q", "Chat about coding", conversation.Title)
	}
	if conversation.ID == "" {
		t.Error("expected conversation to have a non-empty ID")
	}
	if conversation.CreatedAt.IsZero() {
		t.Error("expected conversation to have a CreatedAt timestamp")
	}
}

func TestCreateConversation_InvalidBody(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/conversations", `{invalid json}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
}

func TestCreateConversation_MissingRequiredTitle(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/conversations", `{}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
	if len(env.Error.Fields) == 0 {
		t.Error("expected validation field errors in response")
	}
}

// --------------------------------------------------------------------------
// TestGetConversations
// --------------------------------------------------------------------------

func TestGetConversations_Empty(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/conversations", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var conversations []models.Conversation
	_, err = DecodeResponseData(resp, &conversations)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(conversations) != 0 {
		t.Errorf("expected empty list, got %d items", len(conversations))
	}
}

func TestGetConversations_WithItems(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two conversations directly into the repository.
	_, err := repositories.CreateConversation(models.Conversation{Title: "Chat A"})
	if err != nil {
		t.Fatalf("failed to seed conversation: %v", err)
	}
	_, err = repositories.CreateConversation(models.Conversation{Title: "Chat B"})
	if err != nil {
		t.Fatalf("failed to seed conversation: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/conversations", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var conversations []models.Conversation
	_, err = DecodeResponseData(resp, &conversations)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(conversations) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(conversations))
	}
	// Verify titles (order is not guaranteed).
	titles := map[string]bool{"Chat A": false, "Chat B": false}
	for _, c := range conversations {
		titles[c.Title] = true
	}
	for title, found := range titles {
		if !found {
			t.Errorf("expected conversation %q in results", title)
		}
	}
}

// --------------------------------------------------------------------------
// TestGetConversation
// --------------------------------------------------------------------------

func TestGetConversation_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	original, err := repositories.CreateConversation(models.Conversation{Title: "My Chat"})
	if err != nil {
		t.Fatalf("failed to seed conversation: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/conversations/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var conversation models.Conversation
	_, err = DecodeResponseData(resp, &conversation)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if conversation.ID != original.ID {
		t.Errorf("expected ID %q, got %q", original.ID, conversation.ID)
	}
	if conversation.Title != original.Title {
		t.Errorf("expected title %q, got %q", original.Title, conversation.Title)
	}
}

func TestGetConversation_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/conversations/nonexistent1", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
}

// --------------------------------------------------------------------------
// TestUpdateConversation
// --------------------------------------------------------------------------

func TestUpdateConversation_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	original, err := repositories.CreateConversation(models.Conversation{Title: "Old title"})
	if err != nil {
		t.Fatalf("failed to seed conversation: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/conversations/"+string(original.ID), `{"title":"Updated title"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var conversation models.Conversation
	_, err = DecodeResponseData(resp, &conversation)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if conversation.Title != "Updated title" {
		t.Errorf("expected title %q, got %q", "Updated title", conversation.Title)
	}
}

func TestUpdateConversation_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/conversations/nonexistent1", `{"title":"Should fail"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
}

// --------------------------------------------------------------------------
// TestDeleteConversation
// --------------------------------------------------------------------------

func TestDeleteConversation_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	original, err := repositories.CreateConversation(models.Conversation{Title: "Deletable"})
	if err != nil {
		t.Fatalf("failed to seed conversation: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/conversations/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify the conversation is actually gone.
	_, err = repositories.GetConversation(models.Conversation{Base: models.Base{ID: original.ID}})
	if err == nil {
		t.Error("expected error fetching deleted conversation, got nil")
	}
}

func TestDeleteConversation_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/conversations/nonexistent1", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
}
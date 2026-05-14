package controllers_test

import (
	"net/http"
	"testing"

	"jarvis/models"
	"jarvis/repositories"
)

// --------------------------------------------------------------------------
// TestCreateNote
// --------------------------------------------------------------------------

func TestCreateNote_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/notes", `{"title":"Meeting notes","content":"Discuss project roadmap"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var note models.Note
	err = DecodeJSON(resp, &note)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if note.Title != "Meeting notes" {
		t.Errorf("expected title %q, got %q", "Meeting notes", note.Title)
	}
	if note.Content != "Discuss project roadmap" {
		t.Errorf("expected content %q, got %q", "Discuss project roadmap", note.Content)
	}
	if note.ID == "" {
		t.Error("expected note to have a non-empty ID")
	}
	if note.CreatedAt.IsZero() {
		t.Error("expected note to have a CreatedAt timestamp")
	}
}

func TestCreateNote_InvalidBody(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/notes", `{invalid json}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
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

func TestCreateNote_MissingRequiredTitle(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/notes", `{"content":"missing title"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	var errResp struct {
		Error  string `json:"error"`
		Fields []any  `json:"fields"`
	}
	err = DecodeJSON(resp, &errResp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response")
	}
}

// --------------------------------------------------------------------------
// TestGetNotes
// --------------------------------------------------------------------------

func TestGetNotes_Empty(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/notes", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var notes []models.Note
	err = DecodeJSON(resp, &notes)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("expected empty list, got %d items", len(notes))
	}
}

func TestGetNotes_WithItems(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two notes directly into the repository.
	_, err := repositories.CreateNote(models.Note{Title: "Note A", Content: "First note"})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}
	_, err = repositories.CreateNote(models.Note{Title: "Note B", Content: "Second note"})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/notes", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var notes []models.Note
	err = DecodeJSON(resp, &notes)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
	// Verify titles (order is not guaranteed, so check both exist).
	titles := map[string]bool{"Note A": false, "Note B": false}
	for _, note := range notes {
		titles[note.Title] = true
	}
	for title, found := range titles {
		if !found {
			t.Errorf("expected note %q in results", title)
		}
	}
}

// --------------------------------------------------------------------------
// TestGetNote
// --------------------------------------------------------------------------

func TestGetNote_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a note so we know its ID.
	original, err := repositories.CreateNote(models.Note{Title: "Single note", Content: "Detail"})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/notes/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var note models.Note
	err = DecodeJSON(resp, &note)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if note.ID != original.ID {
		t.Errorf("expected ID %q, got %q", original.ID, note.ID)
	}
	if note.Title != original.Title {
		t.Errorf("expected title %q, got %q", original.Title, note.Title)
	}
	if note.Content != original.Content {
		t.Errorf("expected content %q, got %q", original.Content, note.Content)
	}
}

func TestGetNote_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/notes/nonexistent1", "")
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

// --------------------------------------------------------------------------
// TestUpdateNote
// --------------------------------------------------------------------------

func TestUpdateNote_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a note.
	original, err := repositories.CreateNote(models.Note{Title: "Old title", Content: "Old content"})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/notes/"+string(original.ID), `{"title":"Updated title","content":"Updated content"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var note models.Note
	err = DecodeJSON(resp, &note)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if note.Title != "Updated title" {
		t.Errorf("expected title %q, got %q", "Updated title", note.Title)
	}
	if note.Content != "Updated content" {
		t.Errorf("expected content %q, got %q", "Updated content", note.Content)
	}
}

func TestUpdateNote_PartialUpdate(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a note.
	original, err := repositories.CreateNote(models.Note{Title: "Original", Content: "Original content"})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	app := newTestApp()

	// Update only the title.
	resp, err := PerformRequest(app, "PATCH", "/v1/notes/"+string(original.ID), `{"title":"Only title changed"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var note models.Note
	err = DecodeJSON(resp, &note)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if note.Title != "Only title changed" {
		t.Errorf("expected title %q, got %q", "Only title changed", note.Title)
	}
	// Content should remain unchanged since we didn't send it.
	if note.Content != "Original content" {
		t.Errorf("expected content %q, got %q", "Original content", note.Content)
	}
}

func TestUpdateNote_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/notes/nonexistent1", `{"title":"Should fail"}`)
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

// --------------------------------------------------------------------------
// TestDeleteNote
// --------------------------------------------------------------------------

func TestDeleteNote_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a note.
	original, err := repositories.CreateNote(models.Note{Title: "Deletable", Content: "Will be deleted"})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/notes/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	// Verify the note is actually gone.
	_, err = repositories.GetNote(models.Note{Base: models.Base{ID: original.ID}})
	if err == nil {
		t.Error("expected error fetching deleted note, got nil")
	}
}

func TestDeleteNote_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/notes/nonexistent1", "")
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
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
	_, err = DecodeResponseData(resp, &note)
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

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
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

func TestCreateNote_WithNewTags(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	body := `{"title":"Notes with tags","content":"Has tags","tags":[{"name":"urgent"},{"name":"personal"}]}`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var note models.Note
	_, err = DecodeResponseData(resp, &note)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(note.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(note.Tags))
	}

	// Verify tag names and IDs.
	names := make(map[string]bool)
	for _, tag := range note.Tags {
		names[tag.Name] = true
		if tag.ID == "" {
			t.Errorf("tag %q has empty ID", tag.Name)
		}
	}
	if !names["urgent"] {
		t.Error("expected tag 'urgent' in response")
	}
	if !names["personal"] {
		t.Error("expected tag 'personal' in response")
	}

	// Verify the tags actually exist in the database.
	tags, err := repositories.GetTags(0, 0)
	if err != nil {
		t.Fatalf("failed to get tags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags in DB, got %d", len(tags))
	}
}

func TestCreateNote_ReusesExistingTagByID(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Create a tag first.
	existingTag, err := repositories.CreateTag(models.Tag{Name: "existing-tag"})
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	app := newTestApp()

	// Create a note referencing the existing tag by its ID.
	body := `{"title":"Note with existing tag","content":"Using existing tag","tags":[{"id":"` + string(existingTag.ID) + `","name":"existing-tag"},{"name":"new-tag"}]}`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var note models.Note
	_, err = DecodeResponseData(resp, &note)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(note.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(note.Tags))
	}

	// Verify the existing tag is not duplicated (only 2 tags total in DB).
	tags, err := repositories.GetTags(0, 0)
	if err != nil {
		t.Fatalf("failed to get tags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected exactly 2 tags in DB (existing + new), got %d", len(tags))
	}
}

// TestCreateNote_WithTags_DeduplicatesByName verifies that when two notes
// share the same tag name, the second creation silently reuses the existing
// tag (the unique constraint on `name` prevents duplicate Tag rows, and
// GORM's many2many creates only one tag record).
func TestCreateNote_WithTags_DeduplicatesByName(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	// First note with tag "shared".
	body1 := `{"title":"First","content":"First","tags":[{"name":"shared"}]}`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body1)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for first note, got %d", resp.StatusCode)
	}

	// Second note with the same tag name "shared". GORM's Create inserts the
	// tag first; the unique constraint prevents a duplicate Tag row, but
	// GORM silently handles the conflict — the note is still created with
	// status 201.
	body2 := `{"title":"Second","content":"Second","tags":[{"name":"shared"}]}`
	resp, err = PerformRequest(app, "POST", "/v1/notes", body2)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for second note, got %d", resp.StatusCode)
	}

	// Only 1 tag in the database (the unique constraint prevented duplicates).
	tags, err := repositories.GetTags(0, 0)
	if err != nil {
		t.Fatalf("failed to get tags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected exactly 1 tag in DB (unique constraint prevented duplicates), got %d", len(tags))
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
	_, err = DecodeResponseData(resp, &notes)
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
	_, err = DecodeResponseData(resp, &notes)
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
	_, err = DecodeResponseData(resp, &note)
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

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
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
	_, err = DecodeResponseData(resp, &note)
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
	_, err = DecodeResponseData(resp, &note)
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

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
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

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
}

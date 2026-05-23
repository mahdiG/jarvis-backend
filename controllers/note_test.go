package controllers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"jarvis/models"
	"jarvis/repositories"
)

// --------------------------------------------------------------------------
// TestCreateNotes
// --------------------------------------------------------------------------

func TestCreateNotes_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	body := `[{"title":"Meeting notes","content":"Discuss project roadmap"},{"title":"Shopping list","content":"Buy groceries"}]`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var notes []models.Note
	_, err = DecodeResponseData(resp, &notes)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
	if notes[0].Title != "Meeting notes" {
		t.Errorf("expected title %q, got %q", "Meeting notes", notes[0].Title)
	}
	if notes[1].Title != "Shopping list" {
		t.Errorf("expected title %q, got %q", "Shopping list", notes[1].Title)
	}
	if notes[0].ID == "" {
		t.Error("expected note to have a non-empty ID")
	}
}

func TestCreateNotes_InvalidBody(t *testing.T) {
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

func TestCreateNotes_MissingRequiredTitle(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	// One note without a title should fail validation.
	body := `[{"content":"missing title"}]`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body)
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

func TestCreateNotes_WithNewTags(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	body := `[{"title":"Notes with tags","content":"Has tags","tags":[{"name":"urgent"},{"name":"personal"}]}]`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var notes []models.Note
	_, err = DecodeResponseData(resp, &notes)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if len(notes[0].Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(notes[0].Tags))
	}

	// Verify tag names and IDs.
	names := make(map[string]bool)
	for _, tag := range notes[0].Tags {
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

func TestCreateNotes_ReusesExistingTagByID(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Create a tag first.
	existingTag, err := repositories.CreateTag(models.Tag{Name: "existing-tag"})
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	app := newTestApp()

	// Create a note referencing the existing tag by its ID.
	body := `[{"title":"Note with existing tag","content":"Using existing tag","tags":[{"id":"` + string(existingTag.ID) + `","name":"existing-tag"},{"name":"new-tag"}]}]`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var notes []models.Note
	_, err = DecodeResponseData(resp, &notes)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if len(notes[0].Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(notes[0].Tags))
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

// TestCreateNotes_DeduplicatesByName verifies that when two notes share the
// same tag name within a single batch, the tag rows are deduplicated
// (the unique constraint prevents duplicate Tag rows).
func TestCreateNotes_DeduplicatesByName(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	// Create two notes sharing the same tag name in a single batch.
	body := `[{"title":"First","content":"First","tags":[{"name":"shared"}]},{"title":"Second","content":"Second","tags":[{"name":"shared"}]}]`
	resp, err := PerformRequest(app, "POST", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
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
	_, err := repositories.CreateNotes([]models.Note{
		{Title: "Note A", Content: "First note"},
		{Title: "Note B", Content: "Second note"},
	})
	if err != nil {
		t.Fatalf("failed to seed notes: %v", err)
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
	seeded, err := repositories.CreateNotes([]models.Note{{Title: "Single note", Content: "Detail"}})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}
	original := seeded[0]

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
// TestUpdateNotes
// --------------------------------------------------------------------------

func TestUpdateNotes_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two notes.
	seeded, err := repositories.CreateNotes([]models.Note{
		{Title: "First", Content: "Old content"},
		{Title: "Second", Content: "Also old"},
	})
	if err != nil {
		t.Fatalf("failed to seed notes: %v", err)
	}

	app := newTestApp()

	body := `[{"id":"` + string(seeded[0].ID) + `","title":"Updated first","content":"New content"},{"id":"` + string(seeded[1].ID) + `","title":"Updated second"}]`
	resp, err := PerformRequest(app, "PATCH", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify notes were updated.
	note1, err := repositories.GetNote(models.Note{Base: models.Base{ID: seeded[0].ID}})
	if err != nil {
		t.Fatalf("failed to fetch note 1: %v", err)
	}
	if note1.Title != "Updated first" {
		t.Errorf("expected title %q, got %q", "Updated first", note1.Title)
	}
	if note1.Content != "New content" {
		t.Errorf("expected content %q, got %q", "New content", note1.Content)
	}

	note2, err := repositories.GetNote(models.Note{Base: models.Base{ID: seeded[1].ID}})
	if err != nil {
		t.Fatalf("failed to fetch note 2: %v", err)
	}
	if note2.Title != "Updated second" {
		t.Errorf("expected title %q, got %q", "Updated second", note2.Title)
	}
	// Content should remain unchanged.
	if note2.Content != "Also old" {
		t.Errorf("expected content %q, got %q", "Also old", note2.Content)
	}
}

func TestUpdateNotes_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	// Try updating a non-existent note.
	body := `[{"id":"nonexistent1","title":"Should fail"}]`
	resp, err := PerformRequest(app, "PATCH", "/v1/notes", body)
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
// TestSoftDeleteNotes
// --------------------------------------------------------------------------

func TestSoftDeleteNotes_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two notes.
	seeded, err := repositories.CreateNotes([]models.Note{
		{Title: "Deletable A", Content: "Will be deleted"},
		{Title: "Deletable B", Content: "Also deleted"},
	})
	if err != nil {
		t.Fatalf("failed to seed notes: %v", err)
	}

	app := newTestApp()

	body := `["` + string(seeded[0].ID) + `","` + string(seeded[1].ID) + `"]`
	resp, err := PerformRequest(app, "DELETE", "/v1/notes", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify the notes are gone from normal queries.
	_, err = repositories.GetNote(models.Note{Base: models.Base{ID: seeded[0].ID}})
	if err == nil {
		t.Error("expected error fetching deleted note, got nil")
	}
	_, err = repositories.GetNote(models.Note{Base: models.Base{ID: seeded[1].ID}})
	if err == nil {
		t.Error("expected error fetching deleted note, got nil")
	}

	// Verify they are in the trash.
	trash, err := repositories.GetTrashNotes(0, 0)
	if err != nil {
		t.Fatalf("failed to get trash: %v", err)
	}
	if len(trash) != 2 {
		t.Errorf("expected 2 trashed notes, got %d", len(trash))
	}
}

func TestSoftDeleteNotes_InvalidBody(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/notes", `{invalid}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

// --------------------------------------------------------------------------
// TestGetTrashNotes
// --------------------------------------------------------------------------

func TestGetTrashNotes_Empty(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/notes/trash", "")
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

func TestGetTrashNotes_WithItems(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two notes and soft-delete them.
	seeded, err := repositories.CreateNotes([]models.Note{
		{Title: "Trash A", Content: "First trashed"},
		{Title: "Trash B", Content: "Second trashed"},
	})
	if err != nil {
		t.Fatalf("failed to seed notes: %v", err)
	}

	ids := []models.UID{seeded[0].ID, seeded[1].ID}
	_ = repositories.SoftDeleteNotes(ids)

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/notes/trash", "")
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

	titles := map[string]bool{"Trash A": false, "Trash B": false}
	for _, note := range notes {
		titles[note.Title] = true
	}
	for title, found := range titles {
		if !found {
			t.Errorf("expected note %q in trash results", title)
		}
	}
}

func TestGetTrashNotes_ExcludesActiveNotes(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed one active note and one trashed note.
	seeded, err := repositories.CreateNotes([]models.Note{
		{Title: "Active", Content: "Not deleted"},
		{Title: "Trashed", Content: "Soft-deleted"},
	})
	if err != nil {
		t.Fatalf("failed to seed notes: %v", err)
	}

	_ = repositories.SoftDeleteNotes([]models.UID{seeded[1].ID})

	// Ensure the active note still exists.
	_, err = repositories.GetNote(models.Note{Base: models.Base{ID: seeded[0].ID}})
	if err != nil {
		t.Fatalf("active note should still exist: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/notes/trash", "")
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
	if len(notes) != 1 {
		t.Fatalf("expected 1 trashed note, got %d", len(notes))
	}
	if notes[0].Title != "Trashed" {
		t.Errorf("expected title %q, got %q", "Trashed", notes[0].Title)
	}
}

// --------------------------------------------------------------------------
// TestRestoreNotes
// --------------------------------------------------------------------------

func TestRestoreNotes_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two notes and soft-delete them.
	seeded, err := repositories.CreateNotes([]models.Note{
		{Title: "Restorable A", Content: "Will be restored"},
		{Title: "Restorable B", Content: "Also restored"},
	})
	if err != nil {
		t.Fatalf("failed to seed notes: %v", err)
	}

	ids := []models.UID{seeded[0].ID, seeded[1].ID}
	_ = repositories.SoftDeleteNotes(ids)

	// Verify they're in trash.
	trashNotes, err := repositories.GetTrashNotes(0, 0)
	if err != nil {
		t.Fatalf("failed to get trash: %v", err)
	}
	if len(trashNotes) != 2 {
		t.Fatalf("expected 2 trashed notes, got %d", len(trashNotes))
	}

	app := newTestApp()

	body := `["` + string(seeded[0].ID) + `","` + string(seeded[1].ID) + `"]`
	resp, err := PerformRequest(app, "PATCH", "/v1/notes/trash", body)
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
		t.Fatalf("expected 2 restored notes, got %d", len(notes))
	}

	// Verify they're back from trash.
	trashNotes, err = repositories.GetTrashNotes(0, 0)
	if err != nil {
		t.Fatalf("failed to get trash: %v", err)
	}
	if len(trashNotes) != 0 {
		t.Errorf("expected 0 trashed notes after restore, got %d", len(trashNotes))
	}

	// Verify they're fetchable via normal GetNote.
	for _, seed := range seeded {
		_, err = repositories.GetNote(models.Note{Base: models.Base{ID: seed.ID}})
		if err != nil {
			t.Errorf("expected restored note %q to be fetchable: %v", seed.ID, err)
		}
	}
}

func TestRestoreNotes_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/notes/trash", `["nonexistent1","nonexistent2"]`)
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

func TestRestoreNotes_ActiveNoteNotInTrash(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed an active note (not deleted).
	seeded, err := repositories.CreateNotes([]models.Note{{Title: "Active", Content: "Not deleted"}})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	app := newTestApp()

	body := `["` + string(seeded[0].ID) + `"]`
	resp, err := PerformRequest(app, "PATCH", "/v1/notes/trash", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// Should return 404 because it's not in the trash.
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404 for active note restore, got %d", resp.StatusCode)
	}
}

// --------------------------------------------------------------------------
// TestHardDeleteNotes
// --------------------------------------------------------------------------

func TestHardDeleteNotes_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two notes, soft-delete them, then permanently delete them.
	seeded, err := repositories.CreateNotes([]models.Note{
		{Title: "Permanent A", Content: "Will be permanently deleted"},
		{Title: "Permanent B", Content: "Also gone"},
	})
	if err != nil {
		t.Fatalf("failed to seed notes: %v", err)
	}

	_ = repositories.SoftDeleteNotes([]models.UID{seeded[0].ID, seeded[1].ID})

	app := newTestApp()

	body := `["` + string(seeded[0].ID) + `","` + string(seeded[1].ID) + `"]`
	resp, err := PerformRequest(app, "DELETE", "/v1/notes/trash", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Decode into json.RawMessage to check the data is null.
	var rawData json.RawMessage
	env, err := DecodeResponseData(resp, &rawData)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if string(env.Data) != "null" {
		t.Errorf("expected null data in response, got %s", string(env.Data))
	}

	// Verify they're gone from trash.
	trashNotes, err := repositories.GetTrashNotes(0, 0)
	if err != nil {
		t.Fatalf("failed to get trash: %v", err)
	}
	if len(trashNotes) != 0 {
		t.Errorf("expected 0 trashed notes after permanent delete, got %d", len(trashNotes))
	}
}

func TestHardDeleteNotes_ActiveNotes(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed an active note (not deleted).
	seeded, err := repositories.CreateNotes([]models.Note{{Title: "Active", Content: "Active note"}})
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	app := newTestApp()

	body := `["` + string(seeded[0].ID) + `"]`
	resp, err := PerformRequest(app, "DELETE", "/v1/notes/trash", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// HardDelete works on any note regardless of soft-delete status.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify it's gone from the DB entirely.
	notes, err := repositories.GetNotes(0, 0)
	if err != nil {
		t.Fatalf("failed to get notes: %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("expected 0 notes after hard delete, got %d", len(notes))
	}

	// Verify it's not in trash either.
	trashNotes, err := repositories.GetTrashNotes(0, 0)
	if err != nil {
		t.Fatalf("failed to get trash: %v", err)
	}
	if len(trashNotes) != 0 {
		t.Errorf("expected 0 trashed notes, got %d", len(trashNotes))
	}
}

func TestHardDeleteNotes_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	// HardDelete on non-existent IDs does not error (Delete with no matching
	// rows is Ok in GORM).
	resp, err := PerformRequest(app, "DELETE", "/v1/notes/trash", `["nonexistent1"]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
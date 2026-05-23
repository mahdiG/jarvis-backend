package controllers_test

import (
	"net/http"
	"testing"

	"jarvis/models"
	"jarvis/repositories"
)

// --------------------------------------------------------------------------
// TestCreateTags
// --------------------------------------------------------------------------

func TestCreateTags_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tags", `[{"name":"urgent"},{"name":"personal"}]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var tags []models.Tag
	_, err = DecodeResponseData(resp, &tags)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	names := map[string]bool{"urgent": false, "personal": false}
	for _, tag := range tags {
		if tag.ID == "" {
			t.Error("expected each tag to have a non-empty ID")
		}
		if tag.CreatedAt.IsZero() {
			t.Error("expected each tag to have a CreatedAt timestamp")
		}
		names[tag.Name] = true
	}
	for name, found := range names {
		if !found {
			t.Errorf("expected tag %q in results", name)
		}
	}
}

func TestCreateTags_InvalidBody(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tags", `{invalid json}`)
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

func TestCreateTags_MissingRequiredName(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tags", `[{}]`)
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
// TestGetTags
// --------------------------------------------------------------------------

func TestGetTags_Empty(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tags", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tags []models.Tag
	_, err = DecodeResponseData(resp, &tags)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected empty list, got %d items", len(tags))
	}
}

func TestGetTags_WithItems(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two tags directly into the repository.
	_, err := repositories.CreateTag(models.Tag{Name: "urgent"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}
	_, err = repositories.CreateTag(models.Tag{Name: "personal"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tags", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tags []models.Tag
	_, err = DecodeResponseData(resp, &tags)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	// Verify names (order is not guaranteed, so check both exist).
	names := map[string]bool{"urgent": false, "personal": false}
	for _, tag := range tags {
		names[tag.Name] = true
	}
	for name, found := range names {
		if !found {
			t.Errorf("expected tag %q in results", name)
		}
	}
}

// --------------------------------------------------------------------------
// TestGetTag
// --------------------------------------------------------------------------

func TestGetTag_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a tag so we know its ID.
	original, err := repositories.CreateTag(models.Tag{Name: "single-tag"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tags/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tag models.Tag
	_, err = DecodeResponseData(resp, &tag)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if tag.ID != original.ID {
		t.Errorf("expected ID %q, got %q", original.ID, tag.ID)
	}
	if tag.Name != original.Name {
		t.Errorf("expected name %q, got %q", original.Name, tag.Name)
	}
}

func TestGetTag_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tags/nonexistent1", "")
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
// TestUpdateTags
// --------------------------------------------------------------------------

func TestUpdateTags_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two tags.
	tag1, err := repositories.CreateTag(models.Tag{Name: "old-name-1"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}
	tag2, err := repositories.CreateTag(models.Tag{Name: "old-name-2"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}

	app := newTestApp()

	body := `[{"id":"` + string(tag1.ID) + `","name":"new-name-1"},{"id":"` + string(tag2.ID) + `","name":"new-name-2"}]`
	resp, err := PerformRequest(app, "PATCH", "/v1/tags", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify updates.
	updated1, err := repositories.GetTag(models.Tag{Base: models.Base{ID: tag1.ID}})
	if err != nil {
		t.Fatalf("failed to fetch updated tag: %v", err)
	}
	if updated1.Name != "new-name-1" {
		t.Errorf("expected name %q, got %q", "new-name-1", updated1.Name)
	}

	updated2, err := repositories.GetTag(models.Tag{Base: models.Base{ID: tag2.ID}})
	if err != nil {
		t.Fatalf("failed to fetch updated tag: %v", err)
	}
	if updated2.Name != "new-name-2" {
		t.Errorf("expected name %q, got %q", "new-name-2", updated2.Name)
	}
}

func TestUpdateTags_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tags", `[{"id":"nonexistent1","name":"Should fail"}]`)
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
// TestDeleteTags
// --------------------------------------------------------------------------

func TestDeleteTags_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two tags.
	tag1, err := repositories.CreateTag(models.Tag{Name: "deletable-1"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}
	tag2, err := repositories.CreateTag(models.Tag{Name: "deletable-2"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}

	app := newTestApp()

	body := `["` + string(tag1.ID) + `","` + string(tag2.ID) + `"]`
	resp, err := PerformRequest(app, "DELETE", "/v1/tags", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify both tags are gone.
	_, err = repositories.GetTag(models.Tag{Base: models.Base{ID: tag1.ID}})
	if err == nil {
		t.Error("expected error fetching deleted tag 1, got nil")
	}
	_, err = repositories.GetTag(models.Tag{Base: models.Base{ID: tag2.ID}})
	if err == nil {
		t.Error("expected error fetching deleted tag 2, got nil")
	}
}

func TestDeleteTags_MissingIDs(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tags", `["nonexistent1","nonexistent2"]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
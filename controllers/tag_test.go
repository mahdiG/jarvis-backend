package controllers_test

import (
	"net/http"
	"testing"

	"jarvis/models"
	"jarvis/repositories"
)

// --------------------------------------------------------------------------
// TestCreateTag
// --------------------------------------------------------------------------

func TestCreateTag_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tags", `{"name":"urgent"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var tag models.Tag
	err = DecodeJSON(resp, &tag)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if tag.Name != "urgent" {
		t.Errorf("expected name %q, got %q", "urgent", tag.Name)
	}
	if tag.ID == "" {
		t.Error("expected tag to have a non-empty ID")
	}
	if tag.CreatedAt.IsZero() {
		t.Error("expected tag to have a CreatedAt timestamp")
	}
}

func TestCreateTag_InvalidBody(t *testing.T) {
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

	var errResp ResponseError
	err = DecodeJSON(resp, &errResp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response")
	}
}

func TestCreateTag_MissingRequiredName(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tags", `{}`)
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
	err = DecodeJSON(resp, &tags)
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
	err = DecodeJSON(resp, &tags)
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
	err = DecodeJSON(resp, &tag)
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
// TestUpdateTag
// --------------------------------------------------------------------------

func TestUpdateTag_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a tag.
	original, err := repositories.CreateTag(models.Tag{Name: "old-name"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tags/"+string(original.ID), `{"name":"new-name"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tag models.Tag
	err = DecodeJSON(resp, &tag)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if tag.Name != "new-name" {
		t.Errorf("expected name %q, got %q", "new-name", tag.Name)
	}
}

func TestUpdateTag_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tags/nonexistent1", `{"name":"Should fail"}`)
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
// TestDeleteTag
// --------------------------------------------------------------------------

func TestDeleteTag_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a tag.
	original, err := repositories.CreateTag(models.Tag{Name: "deletable"})
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tags/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	// Verify the tag is actually gone.
	_, err = repositories.GetTag(models.Tag{Base: models.Base{ID: original.ID}})
	if err == nil {
		t.Error("expected error fetching deleted tag, got nil")
	}
}

func TestDeleteTag_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tags/nonexistent1", "")
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
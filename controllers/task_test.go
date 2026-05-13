package controllers_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"jarvis/models"
	"jarvis/repositories"
	"jarvis/router"

	"github.com/gofiber/fiber/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDB is the shared in-memory SQLite database used by all tests.
var TestDB *gorm.DB

// TestMain initializes the in-memory test database once and runs all tests.
func TestMain(m *testing.M) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	err = db.AutoMigrate(&models.Task{}, &models.Conversation{}, &models.Message{})
	if err != nil {
		panic("failed to migrate test database: " + err.Error())
	}

	TestDB = db
	repositories.Init(db)

	m.Run()
}

// newTestApp creates a new Fiber application with routes registered via the router package.
func newTestApp() *fiber.App {
	app := fiber.New()
	router.Setup(app)
	return app
}

// BeginTx starts a new database transaction and sets it as the global repository DB
// so that all repository calls during the test execute within the transaction.
// Returns a cleanup function that rolls back the transaction and restores the original DB.
func BeginTx() func() {
	tx := TestDB.Begin()
	repositories.Init(tx)
	return func() {
		tx.Rollback()
		repositories.Init(TestDB)
	}
}

// PerformRequest sends an HTTP request to the given Fiber app and returns the response.
func PerformRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	req, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}

// DecodeJSON decodes the JSON response body into the given target.
func DecodeJSON(resp *http.Response, target any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// ResponseError is the standard error response envelope.
type ResponseError struct {
	Error string `json:"error"`
}

// --------------------------------------------------------------------------
// TestCreateTask
// --------------------------------------------------------------------------

func TestCreateTask_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tasks", `{"title":"Buy milk","description":"Remember to buy milk"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var task models.Task
	err = DecodeJSON(resp, &task)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if task.Title != "Buy milk" {
		t.Errorf("expected title %q, got %q", "Buy milk", task.Title)
	}
	if task.Description != "Remember to buy milk" {
		t.Errorf("expected description %q, got %q", "Remember to buy milk", task.Description)
	}
	if task.ID == "" {
		t.Error("expected task to have a non-empty ID")
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected task to have a CreatedAt timestamp")
	}
}

func TestCreateTask_InvalidBody(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tasks", `{invalid json}`)
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

func TestCreateTask_MissingRequiredTitle(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tasks", `{"description":"missing title"}`)
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
// TestGetTasks
// --------------------------------------------------------------------------

func TestGetTasks_Empty(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tasks", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tasks []models.Task
	err = DecodeJSON(resp, &tasks)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty list, got %d items", len(tasks))
	}
}

func TestGetTasks_WithItems(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed two tasks directly into the repository.
	_, err := repositories.CreateTask(models.Task{Title: "Task A", Description: "First task"})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	_, err = repositories.CreateTask(models.Task{Title: "Task B", Description: "Second task"})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tasks", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tasks []models.Task
	err = DecodeJSON(resp, &tasks)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	// Verify titles (order is not guaranteed, so check both exist).
	titles := map[string]bool{"Task A": false, "Task B": false}
	for _, task := range tasks {
		titles[task.Title] = true
	}
	for title, found := range titles {
		if !found {
			t.Errorf("expected task %q in results", title)
		}
	}
}

// --------------------------------------------------------------------------
// TestGetTask
// --------------------------------------------------------------------------

func TestGetTask_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a task so we know its ID.
	original, err := repositories.CreateTask(models.Task{Title: "Single task", Description: "Detail"})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tasks/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var task models.Task
	err = DecodeJSON(resp, &task)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if task.ID != original.ID {
		t.Errorf("expected ID %q, got %q", original.ID, task.ID)
	}
	if task.Title != original.Title {
		t.Errorf("expected title %q, got %q", original.Title, task.Title)
	}
	if task.Description != original.Description {
		t.Errorf("expected description %q, got %q", original.Description, task.Description)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tasks/nonexistent1", "")
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
// TestUpdateTask
// --------------------------------------------------------------------------

func TestUpdateTask_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a task.
	original, err := repositories.CreateTask(models.Task{Title: "Old title", Description: "Old description"})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tasks/"+string(original.ID), `{"title":"Updated title","description":"Updated description"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var task models.Task
	err = DecodeJSON(resp, &task)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if task.Title != "Updated title" {
		t.Errorf("expected title %q, got %q", "Updated title", task.Title)
	}
	if task.Description != "Updated description" {
		t.Errorf("expected description %q, got %q", "Updated description", task.Description)
	}
}

func TestUpdateTask_PartialUpdate(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a task.
	original, err := repositories.CreateTask(models.Task{Title: "Original", Description: "Original description"})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	app := newTestApp()

	// Update only the title.
	resp, err := PerformRequest(app, "PATCH", "/v1/tasks/"+string(original.ID), `{"title":"Only title changed"}`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var task models.Task
	err = DecodeJSON(resp, &task)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if task.Title != "Only title changed" {
		t.Errorf("expected title %q, got %q", "Only title changed", task.Title)
	}
	// Description should remain unchanged since we didn't send it.
	if task.Description != "Original description" {
		t.Errorf("expected description %q, got %q", "Original description", task.Description)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tasks/nonexistent1", `{"title":"Should fail"}`)
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
// TestDeleteTask
// --------------------------------------------------------------------------

func TestDeleteTask_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a task.
	original, err := repositories.CreateTask(models.Task{Title: "Deletable", Description: "Will be deleted"})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tasks/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	// Verify the task is actually gone.
	_, err = repositories.GetTask(models.Task{Base: models.Base{ID: original.ID}})
	if err == nil {
		t.Error("expected error fetching deleted task, got nil")
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tasks/nonexistent1", "")
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

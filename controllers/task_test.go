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

	err = db.AutoMigrate(&models.Task{}, &models.Note{}, &models.Conversation{}, &models.Message{}, &models.Tag{})
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

// ResponseEnvelope is the standard API response envelope used by tests
// to decode responses wrapped in controllers.Response.
type ResponseEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Message string          `json:"message"`
		Fields  json.RawMessage `json:"fields"`
	} `json:"error"`
}

// DecodeResponseData decodes the full response envelope and extracts the data
// field into the given target. Returns the full envelope so callers can inspect
// error/meta if needed.
func DecodeResponseData(resp *http.Response, target any) (*ResponseEnvelope, error) {
	var env ResponseEnvelope
	if err := DecodeJSON(resp, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return &env, nil
	}
	if err := json.Unmarshal(env.Data, target); err != nil {
		return &env, err
	}
	return &env, nil
}

// ResponseError is the legacy error envelope (for validate handler errors that aren't wrapped).
type ResponseError struct {
	Error string `json:"error"`
}

// --------------------------------------------------------------------------
// TestCreateTasks
// --------------------------------------------------------------------------

func TestCreateTasks_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tasks", `[{"title":"Buy milk","description":"Remember to buy milk"}]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var tasks []models.Task
	_, err = DecodeResponseData(resp, &tasks)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Buy milk" {
		t.Errorf("expected title %q, got %q", "Buy milk", tasks[0].Title)
	}
	if tasks[0].Description != "Remember to buy milk" {
		t.Errorf("expected description %q, got %q", "Remember to buy milk", tasks[0].Description)
	}
	if tasks[0].ID == "" {
		t.Error("expected task to have a non-empty ID")
	}
	if tasks[0].CreatedAt.IsZero() {
		t.Error("expected task to have a CreatedAt timestamp")
	}
}

func TestCreateTasks_Multiple(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tasks", `[{"title":"Task A"},{"title":"Task B"}]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var tasks []models.Task
	_, err = DecodeResponseData(resp, &tasks)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestCreateTasks_InvalidBody(t *testing.T) {
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

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
}

func TestCreateTasks_MissingRequiredTitle(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "POST", "/v1/tasks", `[{"description":"missing title"}]`)
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
	_, err = DecodeResponseData(resp, &tasks)
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
	_, err := repositories.CreateTasks([]models.Task{{Title: "Task A", Description: "First task"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	_, err = repositories.CreateTasks([]models.Task{{Title: "Task B", Description: "Second task"}})
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
	_, err = DecodeResponseData(resp, &tasks)
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
	created, err := repositories.CreateTasks([]models.Task{{Title: "Single task", Description: "Detail"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	original := created[0]

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tasks/"+string(original.ID), "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var task models.Task
	_, err = DecodeResponseData(resp, &task)
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

	env, err := DecodeResponseData(resp, nil)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Error == nil || env.Error.Message == "" {
		t.Error("expected error message in response")
	}
}

// --------------------------------------------------------------------------
// TestUpdateTasks
// --------------------------------------------------------------------------

func TestUpdateTasks_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a task.
	created, err := repositories.CreateTasks([]models.Task{{Title: "Old title", Description: "Old description"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	original := created[0]

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tasks", `[{"id":"`+string(original.ID)+`","title":"Updated title","description":"Updated description"}]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateTasks_PartialUpdate(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a task.
	created, err := repositories.CreateTasks([]models.Task{{Title: "Original", Description: "Original description"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	original := created[0]

	app := newTestApp()

	// Update only the title.
	resp, err := PerformRequest(app, "PATCH", "/v1/tasks", `[{"id":"`+string(original.ID)+`","title":"Only title changed"}]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateTasks_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tasks", `[{"id":"nonexistent1","title":"Should fail"}]`)
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
// TestSoftDeleteTasks
// --------------------------------------------------------------------------

func TestSoftDeleteTasks_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed a task.
	created, err := repositories.CreateTasks([]models.Task{{Title: "Deletable", Description: "Will be deleted"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	original := created[0]

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tasks", `["`+string(original.ID)+`"]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify the task is soft-deleted (should not be found via normal query).
	_, err = repositories.GetTask(models.Task{Base: models.Base{ID: original.ID}})
	if err == nil {
		t.Error("expected error fetching deleted task, got nil")
	}
}

func TestSoftDeleteTasks_NonExistent(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tasks", `["nonexistent1"]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

// --------------------------------------------------------------------------
// TestGetTrashTasks / RestoreTasks / HardDeleteTasks
// --------------------------------------------------------------------------

func TestGetTrashTasks_Empty(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tasks/trash", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tasks []models.Task
	_, err = DecodeResponseData(resp, &tasks)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty trash, got %d items", len(tasks))
	}
}

func TestGetTrashTasks_WithItems(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed and soft-delete a task.
	created, err := repositories.CreateTasks([]models.Task{{Title: "Trashable"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	err = repositories.SoftDeleteTasks([]models.UID{created[0].ID})
	if err != nil {
		t.Fatalf("failed to soft-delete task: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "GET", "/v1/tasks/trash", "")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tasks []models.Task
	_, err = DecodeResponseData(resp, &tasks)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 trashed task, got %d", len(tasks))
	}
	if tasks[0].Title != "Trashable" {
		t.Errorf("expected title %q, got %q", "Trashable", tasks[0].Title)
	}
}

func TestRestoreTasks_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed and soft-delete a task.
	created, err := repositories.CreateTasks([]models.Task{{Title: "Restorable"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	id := created[0].ID
	err = repositories.SoftDeleteTasks([]models.UID{id})
	if err != nil {
		t.Fatalf("failed to soft-delete task: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tasks/trash", `["`+string(id)+`"]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var tasks []models.Task
	_, err = DecodeResponseData(resp, &tasks)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 restored task, got %d", len(tasks))
	}
	if tasks[0].ID != id {
		t.Errorf("expected ID %q, got %q", id, tasks[0].ID)
	}

	// Verify it's no longer in trash.
	tasks, err = repositories.GetTrashTasks(0, 0)
	if err != nil {
		t.Fatalf("failed to get trash tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected trash to be empty after restore, got %d items", len(tasks))
	}
}

func TestRestoreTasks_NotFound(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "PATCH", "/v1/tasks/trash", `["nonexistent1"]`)
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

func TestHardDeleteTasks_Success(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	// Seed and soft-delete a task.
	created, err := repositories.CreateTasks([]models.Task{{Title: "PermanentDelete"}})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	id := created[0].ID
	err = repositories.SoftDeleteTasks([]models.UID{id})
	if err != nil {
		t.Fatalf("failed to soft-delete task: %v", err)
	}

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tasks/trash", `["`+string(id)+`"]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify the task is gone even from unscoped query.
	var count int64
	TestDB.Unscoped().Model(&models.Task{}).Where("id = ?", id).Count(&count)
	if count != 0 {
		t.Errorf("expected task to be permanently deleted, count = %d", count)
	}
}

func TestHardDeleteTasks_NonExistent(t *testing.T) {
	cleanup := BeginTx()
	defer cleanup()

	app := newTestApp()

	resp, err := PerformRequest(app, "DELETE", "/v1/tasks/trash", `["nonexistent1"]`)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
# Adding New APIs — Developer Guide

This document describes the step-by-step process for adding a new REST API resource to the Jarvis backend. It uses the existing **Tasks API** as the canonical reference.

**Prerequisite**: read `docs/CODE-GUIDELINES.md` first — this guide builds on those conventions.

---

## Overview: The 5-Step Workflow

Every new resource follows the same pipeline:

```
Model → Repository → Controller → Router → Tests
            ↕
        (auto-migrate)
```

1. **Model** — define the data structure in `models/`.
2. **Repository** — implement database access in `repositories/`.
3. **Controller** — write HTTP handlers in `controllers/`.
4. **Router** — register routes in `router/router.go`.
5. **Tests** — write controller tests in `controllers/`.

The following sections walk through each step using a new **Habit** resource as the running example.

---

## Step 1: Model

### Where

`models/habit.go`

### Conventions

- **File name** matches the model name, singular, kebab-case (`habit.go`, not `habits.go`).
- **Embed `Base`** as the first field (provides `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`).
- **Tag order**: `gorm` → `validate`. No `json` tags — the API serializes using Go's default JSON field naming, which preserves the struct field name as-is (CapitalCase). For example, `Title` serializes as `"Title"`, `Description` as `"Description"`.
- **One model per file** — cohesive helpers for the type may live in the same file.
- **Model struct IS the API contract** — no separate request/response DTOs. Add validation tags directly on the model. The request body JSON key names are case-insensitive (Go's `encoding/json` behavior), but responses always use CapitalCase matching the Go field names.

### Example

```go
package models

type Habit struct {
    Base
    Title       string `gorm:"not null" validate:"required"`
    Description string
    StreakCount int    `gorm:"default:0"`
}
```

### What to compare against

See `models/task.go` — identical structure, different fields.

### Migration

Models are auto-migrated in `cmd/api/main.go` via `db.AutoMigrate`. When adding a new model, add it to the `AutoMigrate` call:

```go
// In configs/database.go or main.go — add the new model here
err = db.AutoMigrate(&models.Task{}, &models.Habit{})
```

---

## Step 2: Repository

### Where

`repositories/habit.go`

### Conventions

- **Top-level functions** — no struct or interface. Keep it simple until polymorphism is needed.
- **Pass models by value, return by value** — repositories take `models.Habit` and return `(models.Habit, error)`. Never mutate input pointers.
- **Use `clause.Returning{}`** on Get/Update operations so GORM populates the returned struct with updated DB values.
- **Always check `RowsAffected`** for mutating operations (Update, Delete) — if zero, return `ErrRecordNotFound`.
- **Repository functions are thin** — no business logic, no validation. Just database operations.
- **Database instance** is the package-level `db` variable initialized by `repositories.Init()` (see `repositories/db.go`).

### Standard CRUD pattern

Every resource typically has these five functions:

| Function                              | HTTP Method | Purpose                             |
| ------------------------------------- | ----------- | ----------------------------------- |
| `GetHabits(limit, offset int)`        | GET /       | List all (with optional pagination) |
| `GetHabit(condition models.Habit)`    | GET /:id    | Get one by condition (usually ID)   |
| `CreateHabit(habit models.Habit)`     | POST /      | Create                              |
| `UpdateHabit(habit models.Habit)`     | PATCH /:id  | Update (partial)                    |
| `DeleteHabit(condition models.Habit)` | DELETE /:id | Delete                              |

### Example

```go
package repositories

import (
    "jarvis/models"
    "gorm.io/gorm/clause"
)

func GetHabits(limit, offset int) ([]models.Habit, error) {
    var habits []models.Habit
    query := db.Model(&models.Habit{})
    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }
    result := query.Find(&habits)
    return habits, result.Error
}

func GetHabit(condition models.Habit) (models.Habit, error) {
    var habit models.Habit
    result := db.
        Clauses(clause.Returning{}).
        Where(&condition).
        First(&habit)
    return habit, result.Error
}

func CreateHabit(habit models.Habit) (models.Habit, error) {
    result := db.Create(&habit)
    return habit, result.Error
}

func UpdateHabit(habit models.Habit) (models.Habit, error) {
    result := db.
        Clauses(clause.Returning{}).
        Where("id = ?", habit.ID).
        Updates(&habit)
    if result.Error != nil {
        return habit, result.Error
    }
    if result.RowsAffected == 0 {
        return habit, ErrRecordNotFound
    }
    return habit, nil
}

func DeleteHabit(condition models.Habit) error {
    result := db.
        Where(&condition).
        Delete(&models.Habit{})
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return ErrRecordNotFound
    }
    return nil
}
```

### What to compare against

See `repositories/task.go` — the code is nearly identical; only the model type changes.

### Error sentinels

Use the existing sentinels from `repositories/error.go`:

```go
var ErrRecordNotFound = gorm.ErrRecordNotFound
```

Add new sentinels there as needed (e.g., `ErrDuplicateEntry`, `ErrConflict`).

---

## Step 3: Controller

### Where

`controllers/habit.go`

### Conventions

- **Top-level exported functions** matching the Fiber handler signature `func(c *fiber.Ctx) error`.
- **No controller struct** — keep it simple.
- **Handlers are thin**: parse request → delegate to repository → format response.
- **Use `Validate[T]`** for parsing + validation (see `controllers/validate.go`). This generic helper calls `c.BodyParser` and validates with `go-playground/validator/v10`.
- **Error handling pattern**:
  - Use `errors.Is` to check for `repositories.ErrRecordNotFound` → return 404.
  - Log unexpected errors with `slog.Error`.
  - Return 500 for unexpected database errors.
  - Do NOT leak internal error details to the API consumer.
- **Success response status codes**:
  - GET (list): `200 OK`
  - GET (single): `200 OK`
  - POST: `201 Created`
  - PATCH: `200 OK`
  - DELETE: `200 OK`
- **No separate request DTOs** — the model struct IS the API contract. Parse directly into the model.

### Example

```go
package controllers

import (
    "errors"
    "log/slog"

    "jarvis/models"
    "jarvis/repositories"

    "github.com/gofiber/fiber/v2"
)

func GetHabits(c *fiber.Ctx) error {
    habits, err := repositories.GetHabits(0, 0)
    if err != nil {
        slog.Error("failed to get habits", "error", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to get habits from db",
        })
    }
    return c.JSON(habits)
}

func GetHabit(c *fiber.Ctx) error {
    id := c.Params("id")
    habit, err := repositories.GetHabit(models.Habit{Base: models.Base{ID: models.UID(id)}})
    if err != nil {
        if errors.Is(err, repositories.ErrRecordNotFound) {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "habit not found",
            })
        }
        slog.Error("failed to get habit", "id", id, "error", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to get habit from db",
        })
    }
    return c.JSON(habit)
}

func CreateHabit(c *fiber.Ctx) error {
    var habit models.Habit
    if !Validate(c, &habit) {
        return nil
    }
    habit, err := repositories.CreateHabit(habit)
    if err != nil {
        slog.Error("failed to create habit", "error", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to create habit in db",
        })
    }
    return c.Status(fiber.StatusCreated).JSON(habit)
}

func UpdateHabit(c *fiber.Ctx) error {
    id := c.Params("id")
    var input models.Habit
    if !Validate(c, &input) {
        return nil
    }
    input.ID = models.UID(id)
    habit, err := repositories.UpdateHabit(input)
    if err != nil {
        if errors.Is(err, repositories.ErrRecordNotFound) {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "habit not found",
            })
        }
        slog.Error("failed to update habit", "id", id, "error", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to update habit",
        })
    }
    return c.JSON(habit)
}

func DeleteHabit(c *fiber.Ctx) error {
    id := c.Params("id")
    err := repositories.DeleteHabit(models.Habit{Base: models.Base{ID: models.UID(id)}})
    if err != nil {
        if errors.Is(err, repositories.ErrRecordNotFound) {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "habit not found",
            })
        }
        slog.Error("failed to delete habit", "id", id, "error", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to delete habit",
        })
    }
    return c.SendStatus(fiber.StatusNoContent)
}
```

### What to compare against

See `controllers/task.go` — the controller functions follow an identical structure.

---

## Step 4: Router

### Where

`router/router.go`

### Conventions

- **Versioned API groups**: all endpoints live under `/v1`.
- **Resource groups**: each resource gets its own group (e.g., `/v1/tasks`, `/v1/habits`).
- **RESTful endpoint mapping**:

| Method | Path             | Controller                |
| ------ | ---------------- | ------------------------- |
| GET    | `/v1/habits`     | `controllers.GetHabits`   |
| POST   | `/v1/habits`     | `controllers.CreateHabit` |
| GET    | `/v1/habits/:id` | `controllers.GetHabit`    |
| PATCH  | `/v1/habits/:id` | `controllers.UpdateHabit` |
| DELETE | `/v1/habits/:id` | `controllers.DeleteHabit` |

### Example

```go
func Setup(app *fiber.App) {
    v1 := app.Group("/v1")

    // ... existing routes ...

    habits := v1.Group("/habits")
    habits.Get("/", controllers.GetHabits)
    habits.Post("/", controllers.CreateHabit)
    habits.Get("/:id", controllers.GetHabit)
    habits.Patch("/:id", controllers.UpdateHabit)
    habits.Delete("/:id", controllers.DeleteHabit)
}
```

### What to compare against

See `router/router.go` — the `tasks` group is the template. Add your new group after it, keeping the same structure.

---

## Step 5: Tests

### Where

`controllers/habit_test.go`

### Conventions

- **Same package convention**: use `package controllers_test` for black-box testing (testing the exported API).
- **File name**: `<resource>_test.go` — matches the file being tested.
- **Reuse test infrastructure** from `controllers/task_test.go`:
  - `TestMain` — initializes the in-memory SQLite database once.
  - `BeginTx()` — wraps each test in a transaction that rolls back automatically.
  - `newTestApp()` — creates a fresh Fiber app with all routes registered.
  - `PerformRequest()` — sends HTTP requests to the Fiber app.
  - `DecodeJSON()` — deserializes JSON responses.
  - `ResponseError` — standard error response envelope.

These helpers live in the shared `controllers` test package (the first test file to be compiled will provide them). When adding a new test file, you can call them directly.

### Test naming

`Test[FunctionName]_[Scenario]` — e.g.:

- `TestCreateHabit_Success`
- `TestCreateHabit_InvalidBody`
- `TestCreateHabit_MissingRequiredTitle`
- `TestGetHabits_Empty`
- `TestGetHabits_WithItems`
- `TestGetHabit_Success`
- `TestGetHabit_NotFound`
- `TestUpdateHabit_Success`
- `TestUpdateHabit_PartialUpdate`
- `TestUpdateHabit_NotFound`
- `TestHabit_Success`
- `TestDeleteHabit_NotFound`

### Test structure

Use `// --------------------------------------------------------------------------` section headers to visually separate endpoint groups.

Success path tests should:

- Assert HTTP status code.
- Assert key fields in the response body.
- Assert the object was actually persisted (e.g., for delete, verify fetch returns error).

Error path tests should:

- Assert the correct HTTP error status.
- Assert the error response body has a non-empty `"error"` field.

### Example test

```go
package controllers_test

import (
    "net/http"
    "testing"

    "jarvis/models"
    "jarvis/repositories"
)

// --------------------------------------------------------------------------
// TestCreateHabit
// --------------------------------------------------------------------------

func TestCreateHabit_Success(t *testing.T) {
    cleanup := BeginTx()
    defer cleanup()

    app := newTestApp()

    resp, err := PerformRequest(app, "POST", "/v1/habits", `{"title":"Exercise","description":"Daily workout","streak_count":0}`)
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("expected status 201, got %d", resp.StatusCode)
    }

    var habit models.Habit
    err = DecodeJSON(resp, &habit)
    if err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    if habit.Title != "Exercise" {
        t.Errorf("expected title %q, got %q", "Exercise", habit.Title)
    }
    if habit.StreakCount != 0 {
        t.Errorf("expected streak_count 0, got %d", habit.StreakCount)
    }
    if habit.ID == "" {
        t.Error("expected habit to have a non-empty ID")
    }
    if habit.CreatedAt.IsZero() {
        t.Error("expected habit to have a CreatedAt timestamp")
    }
}

// --------------------------------------------------------------------------
// TestGetHabit
// --------------------------------------------------------------------------

func TestGetHabit_NotFound(t *testing.T) {
    cleanup := BeginTx()
    defer cleanup()

    app := newTestApp()

    resp, err := PerformRequest(app, "GET", "/v1/habits/nonexistent1", "")
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
```

### What to compare against

See `controllers/task_test.go` — this is the canonical reference for the test pattern. Copy the structure and replace "Task" with "Habit".

### If your model adds new fields to AutoMigrate

Add the new model to `AutoMigrate` in `TestMain`:

```go
func TestMain(m *testing.M) {
    // ...
    err = db.AutoMigrate(&models.Task{}, &models.Habit{})
    // ...
}
```

---

## Pattern Reference

### HTTP Method → Controller → Repository Mapping

| HTTP Method   | Controller    | Repository                 | Returns              |
| ------------- | ------------- | -------------------------- | -------------------- |
| `GET /`       | `GetHabits`   | `GetHabits(limit, offset)` | `[]models.Habit`     |
| `GET /:id`    | `GetHabit`    | `GetHabit(condition)`      | `models.Habit`       |
| `POST /`      | `CreateHabit` | `CreateHabit(habit)`       | `models.Habit` (201) |
| `PATCH /:id`  | `UpdateHabit` | `UpdateHabit(habit)`       | `models.Habit`       |
| `DELETE /:id` | `DeleteHabit` | `DeleteHabit(condition)`   | (204)                |

### HTTP Status Codes

| Code                        | When to Use                                                          |
| --------------------------- | -------------------------------------------------------------------- |
| `200 OK`                    | Successful GET (list or single), successful PATCH, successful DELETE |
| `201 Created`               | Successful POST (resource created)                                   |
| `400 Bad Request`           | Invalid JSON body, validation failure                                |
| `404 Not Found`             | Resource not found (`ErrRecordNotFound`)                             |
| `500 Internal Server Error` | Unexpected database or server error                                  |

### Error Sentinels → HTTP Mappings

| Sentinel                         | HTTP Status | Error Message                     |
| -------------------------------- | ----------- | --------------------------------- |
| `repositories.ErrRecordNotFound` | `404`       | `"<resource> not found"`          |
| (unexpected error)               | `500`       | `"failed to <action> <resource>"` |

---

## Common Gotchas

1. **Forgetting to register routes** — a controller is useless until wired in `router/router.go`. Always add both the route group and all endpoint mappings.

2. **Not checking `RowsAffected`** — `Update` and `Delete` must check `result.RowsAffected == 0` and return `ErrRecordNotFound`. Without this, updating/deleting a non-existent record silently succeeds.

3. **Forgetting `clause.Returning{}`** — Without it, Get/Update queries won't return the updated DB values (e.g., auto-set timestamps, defaults). The returned struct may have zero values.

4. **Using pointer receivers incorrectly** — Pass models by value, return by value. Do NOT pass `*models.Habit` and mutate inside the function.

5. **Wrong tag order** — Always `gorm` → `validate`. No `json` tags are used.

6. **Forgetting to add the model to `AutoMigrate`** — The model struct exists in code, but the database table won't be created. This causes silent query failures. Check `cmd/api/main.go` (or wherever `AutoMigrate` is called) when adding a new model.

7. **Skipping tests** — Every CRUD endpoint should have at minimum: one success test and one not-found test. Validation error tests are important for required fields. See `controllers/task_test.go` for the complete pattern.

---

## Checklist for Adding a New API

Before considering a new API done, verify each item:

- [ ] Model file created in `models/` with `Base` embedded and proper tags.
- [ ] Model added to `AutoMigrate` call in startup code.
- [ ] Repository file created in `repositories/` with all CRUD functions.
- [ ] Controller file created in `controllers/` with all HTTP handlers.
- [ ] Routes registered in `router/router.go` under `/v1`.
- [ ] Test file created in `controllers/` with success and error path tests.
- [ ] Application compiles (`go build ./...`).
- [ ] Tests pass (`go test ./...`).

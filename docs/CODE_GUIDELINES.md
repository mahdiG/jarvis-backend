# Jarvis Backend — Code Guidelines

This document is the **source of truth** for how we write code in this repository.

- **Developers**: follow these guidelines when writing or reviewing code.
- **AI agents**: you **must** follow these guidelines when generating or editing code in this repo.
  - If the user gives feedback like "don't do X again" or "prefer Y here", **update this document** (or the most relevant doc in `docs/`) so the mistake is not repeated.

## Principles (highest priority)

- **Clarity over cleverness**: optimize for readability, predictability, and maintainability.
- **Make intent obvious**: write code that a teammate can understand quickly without extra context.
- **Small, composable pieces**: prefer functional and modular code where each function does one clear thing.
- **Stability through tests**: write tests so changing code later is safe and fast.
- **Fail fast**: validate inputs early, return errors explicitly, and don't silently swallow failures.
- **Follow best practices**: Always follow best practices and standards so we have a good code base.
- **Start simple, iterate**: prefer a flat/infinite recursive structure over premature micro-categorization. The task model is a single recursive table — this lets you restructure later without migration pain.

---

## Naming

- **Never use 1-2 letter names or unclear abbreviations** — this is a hard rule, not a suggestion. Bad examples: `h`, `b`, `t`, `r`, `hs`, `res`, `obj`, `srv`, `repo`, `svc`, `data`. A name must be meaningful and obvious even outside of its immediate context. If you show the name to a random developer unfamiliar with the code, they should understand what it refers to without explanation.
- **Prefer descriptive names over abbreviations**: use `CreateHabit` instead of `NewH`, `GetHabitByID` instead of `Get`, `IsAuthenticated` instead of `Auth`. No abbreviation is too obvious to skip — always spell out the meaning.
- **The "random person" test**: can you tell a developer unfamiliar with this codebase what a function/variable does just by reading its name? If not, rename it. `GetTasks` passes — it's obvious. `h`, `b`, `r` fail — they carry no meaning to anyone.
- **Prefer domain language**: name things after product concepts (task, habit, session, streak, event).
- **Boolean naming**:
  - Use `Is/Has/Can/Should` prefixes: `IsAuthenticated`, `HasUnsavedChanges`.
  - Avoid double negatives.
- **Function naming**:
  - Prefer verbs: `CreateTask`, `GetTask`, `DeleteTask`, `ParseSchedule`.
  - Function names must be unambiguous. `Get` is never acceptable — use `GetTask`, `GetActiveGoals`, etc. A function name should make the return value obvious without reading its signature.
- **Case conventions** (Go standard):
  - **PascalCase** for: exported functions (`CreateTask`), exported types (`Task`, `TaskController`), exported constants (`DefaultPort`, `UIDLength`).
  - **camelCase** for: unexported functions (`parseSchedule`), unexported types (`taskRepository`), variables (`completedCount`), local constants.
  - **snake_case** for: JSON field names (`"created_at"`), database column names (`updated_at`), environment variable names (`DATABASE_DSN`).
  - **kebab-case** for: URL paths (`/api/tasks/:id`), middleware names, route group prefixes.
  - Never mix: don't use PascalCase in JSON tags or snake_case in Go identifiers.
- **Directory and file names**:
  - **Directories**: use `kebab-case`. May be singular or plural depending on what reads naturally (e.g., `repositories/`, `controllers/`, `utils/`, `configs/`).
  - **Go source files**: use `kebab-case`. File names are usually singular — they reflect the primary type or responsibility (e.g., `task.go`, `focus-session.go`, `weekly-review.go`, not `tasks.go` or `focus-sessions.go`).
  - **One primary exported type per file**: the file name matches that type. Cohesive helpers for the type may live in the same file. Avoid splitting one type across files or cramming many unrelated types into one file.
- **Package naming**: package names must match their directory name — short, lowercase, no underscores, no mixed casing. For example, `package controllers` lives in `controllers/`, `package router` lives in `router/`. Avoid generic names like `package common` or `package util`.
- **Export discipline**: start with unexported identifiers. Export only when an external package genuinely needs access. This keeps the public surface minimal and intentional. Avoid "export and regret" — unexporting is a breaking change for consumers.
- **Import naming**: rely on Go's default import naming (the last segment of the package path). Use import aliases only to resolve collisions or when the default name is unclear. Never rename imports arbitrarily.
- **Receiver naming**: use meaningful receiver names that match the type. Prefer descriptive abbreviations over single letters (e.g., `task Task`, `service HabitService`). Be consistent across all methods on the same type.
- **File naming for tests**: test files use `_test.go` suffix with the same base name as the file they test (e.g., `task_test.go` for `task.go`). Place in the same package for white-box testing or an `_test` package for black-box testing.

---

## Project structure & package conventions

The codebase follows a layered architecture with clean dependency direction:

```
cmd/api/           # Application entry point (main.go)
configs/           # Initialization, wiring, configuration (e.g., database.go)
constants/         # Shared constants
controllers/       # HTTP handlers (Fiber), request parsing, response formatting — keep thin
docs/              # Documentation (guidelines, vision, progress)
models/            # Data structures, type definitions, validation tags, GORM hooks
repositories/      # Database access and queries via GORM — no business logic
router/            # Route definitions and middleware setup
utils/             # Shared helpers for truly cross-cutting concerns (e.g., slog, pagination)
```

**Layered dependency direction**: `controllers → repositories → models`. Never reverse.

- **No separate DTOs** — model structs ARE the API contract (see [HTTP handler code](#http-handler-code-fiber)).
- **Keep packages focused** — each package has a single responsibility. Don't put database logic in controllers or HTTP parsing in models.



### Model conventions

- **Embed `Base` as the first field** — don't inline ID/timestamps manually.
- **Use GORM tags** on fields: `gorm:"not null"`, `gorm:"index"`, `gorm:"uniqueIndex:idx_name"`.
- **Use `validate` tags** for request validation (`validate:"required"`, `validate:"min=3,max=100"`).
- **Use `json` tags** (snake_case) for JSON serialization.
- **Tag order convention**: `gorm` → `validate` → `json` (data layer first).
- **Domain primitives**: use named types (`type Priority int`, `type Status string`) instead of raw primitives to convey meaning and prevent misuse.
- **One model per file**: file name matches the model name (`task.go` for `Task`).

---

## Repository layer

### Pattern

Repositories use a **package-level `db` variable** initialized by `Init()`:

```go
package repositories

var db *gorm.DB

func Init(database *gorm.DB) {
    db = database
}
```

This is called once at startup in `main.go` and in test setup (`TestMain`).

### Repository function conventions

- **Functions are top-level, not methods on a struct** — no repository struct/interface yet. Keep it simple until polymorphism is actually needed.
- **Pass models by value, return by value** — repositories take `models.Task` (not `*models.Task`) and return `(models.Task, error)`. The caller gets the updated value back. Never mutate input pointers.
- **Return `repositories.ErrRecordNotFound`** when a record is not found or `RowsAffected == 0`. Controllers check this with `errors.Is`.
- **Use `clause.Returning{}`** on Get/Update operations so GORM populates the returned struct with updated DB values if you need the new value.
- **Always check `RowsAffected`** for mutating operations (Update, Delete) to detect missing records.
- **Keep repositories thin** — no business logic, no validation. Just database operations.


### Error sentinels

Defined in `repositories/error.go`:

```go
var ErrRecordNotFound = gorm.ErrRecordNotFound
```

Add more sentinels as needed (`ErrDuplicateEntry`, `ErrConflict`).

### Testing repositories

- Use `repositories.Init(tx)` inside per-test transactions (see [Testing section](#testing)).

---

## Controller layer (HTTP handlers)

### Pattern

- **Handlers are thin**: parse request → delegate to repository → format response.
- **Functions are top-level**, exported, matching the Fiber handler signature: `func(c *fiber.Ctx) error`.
- **No controller struct** — keep it simple until grouping state/logic is warranted.


## Functions, modules, and structure

- **Keep functions short**: if a function is hard to scan, split it.
- **One reason to change**: each function should have a single responsibility.
- **Avoid repetition**:
  - Extract reusable helpers and modules.
  - Prefer shared utilities over copy/paste.
  - If logic is reused in 3+ places, it likely belongs in a shared module.
- **Minimize side effects**: isolate IO (database, network, filesystem) behind repository functions.
- **Pass values, not pointers, to functions that modify state**: do not pass a pointer to a function and mutate the underlying value inside it. Instead, pass the value to the function, let the function modify and return the updated value. This keeps data flow explicit and avoids hidden side effects.

---

## Go & types

- **Don't fight the type system**: prefer specific types over `interface{}` or `any`. Model domain states precisely.
- **Prefer value types over pointers** for small, immutable data (e.g., `UID` as `string`, `Task` as value when not mutated). Use pointers for mutation or optional fields.
- **Use custom types for domain primitives**: use named types (e.g., `type UID string`) instead of raw `string` to convey meaning and prevent misuse.
- **Return concrete types from public functions** unless you need abstraction. Introduce interfaces where callers need polymorphism.
- **Zero-value initialization**: prefer zero values over `nil` where possible (e.g., `var task Task` instead of `nil`).
- **Avoid `panic`** in application code. Only use in truly unrecoverable situations (startup init in `TestMain` is one acceptable case, but prefer `log.Fatalf`).
- **Error handling**:
  - Return errors, don't ignore them. Respect `_ = fn()` only when intentionally discarding a result.
  - Use `fmt.Errorf("context: %w", err)` to wrap errors with context.
  - Use `errors.Is` / `errors.As` for sentinel/wrapped error checking.
  - Define sentinel errors (e.g., `var ErrNotFound = errors.New("not found")`) for expected failure cases.
- **Runtime validation for external data**: anything coming from network (HTTP requests) or storage (DB) is untrusted — validate/parse before use.

---

## Error handling & API responses

- **Fail loudly in development, gracefully in production**:
  - Use structured logging (`slog`) with context for actionable error details.
  - Always respond with proper HTTP status codes and consistent JSON error bodies.
  - Never leak internal error details to the API consumer. Use generic messages for 5xx, specific for 4xx when safe.
- **No silent failures**: never write empty `catch {}` (Go doesn't have it, but avoid unhandled errors).
- **Handler error pattern**:

```go
if err != nil {
    if errors.Is(err, repositories.ErrRecordNotFound) {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "task not found",
        })
    }
    slog.Error("failed to get task", "id", id, "error", err)
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "error": "failed to get task from database",
    })
}
```

## HTTP handler code (Fiber)

- **Handlers should be thin**: parse request, delegate to repository, format response.
- **Use consistent JSON response format**: `{"error": "..."}` for errors, direct object/array for success.
- **Use Fiber's `c.Params`, `c.Query`, `c.BodyParser`** for input extraction.
- **Validate parsed input** before passing to business logic via `ParseAndValidate` (see below).
- **Use `fiber.Map`** for ad-hoc JSON maps, or dedicated response types for complex structures.
- **Use GORM model structs as API contracts directly** — do not create separate request or response DTOs. The `models/` package owns both the database schema and the API contract. Add JSON tags (snake_case) and validation tags directly to the model struct. The frontend sends/receives JSON that maps to these same struct fields. This keeps the API surface in sync with the data layer and eliminates translation code.

---

## Parsing & validation

Use the generic `ParseAndValidate[T]` helper in `controllers/validate.go`:

```go
var task models.Task
if !ParseAndValidate(c, &task) {
    return nil  // response already sent
}
```

This helper:
1. Calls `c.BodyParser(target)` to parse the request body into the target struct.
2. Validates the struct using `go-playground/validator/v10` based on `validate` struct tags.
3. Sends appropriate 400 error responses on failure (parse error or validation errors).
4. Returns `true` on success, `false` on failure (with response already sent).

**Validation error response format** (with details):

```json
{
  "error": "request validation failed",
  "fields": [
    {"field": "Title", "tag": "required", "message": "Key: 'Task.Title' Error:Field validation for 'Title' failed on the 'required' tag"}
  ]
}
```

### Adding validation tags

Add `validate` tags directly to model struct fields:

```go
type Task struct {
    Base
    Title       string `gorm:"not null" validate:"required"`
    Description string
}
```

Available validation tags: `required`, `min`, `max`, `email`, `oneof=active done archived`, etc. See [go-playground/validator docs](https://pkg.go.dev/github.com/go-playground/validator/v10).

---

## Formatting & style

- **Prefer simple control flow**:
  - Early returns over deeply nested `if`s.
  - Guard clauses at the top of functions.
- **Define variables before the `if` statement, not inside it**: declaring a variable inside `if` (`if err := fn(); err != nil {}`) reduces readability and makes it harder to distinguish the condition from the assignment. Always separate:

```go
// Bad
if err := c.BodyParser(&req); err != nil {
    return err
}

// Good
err := c.BodyParser(&req)
if err != nil {
    return err
}
```

- **Be consistent**: match the existing style in the package unless there's a strong reason to change it.
- **Avoid "magic"**:
  - Use constants for shared literals.
  - Document non-obvious constraints in code (with a short comment explaining *why*, not *what*).
- **Use `gofmt` / standard Go formatting** — never fight the formatter.
- **Import ordering**: standard library → third-party → internal modules (jarvis/...), separated by blank lines.
- **Comments**:
  - Prefer explaining *why* a decision was made, not *what* the code does (the code is the *what*).
  - Every exported identifier must have a doc comment (`// PackageName` for packages, `// FunctionName` for functions/types).
  - Unexported helpers that are non-obvious should also have a comment.
  - Keep comments up to date — a stale comment is worse than no comment.
---

## Testing


### Test conventions

- **Write tests for**:
  - Bug fixes (regression tests).
  - Non-trivial business logic (parsing, calculations, state transitions).
  - Each HTTP endpoint (controller tests cover the full request → response cycle).
- **Test file location**: place `_test.go` files in the same package as the code they test (white-box testing) — e.g., `controllers/task_test.go` tests `controllers/task.go`.
- **Use the right test approach**:
  - `testing` package for all tests.
- **Test behavior, not implementation**: assert outcomes and API contract, not internal calls.
- **Keep tests readable**: arrange/act/assert structure, descriptive test names (`TestGetTask_WhenNotFound`).
- **Test data**: use small, focused test helpers to build fixtures. Seed data directly via repository calls in each test (not shared global state):

```go
original, err := repositories.CreateTask(models.Task{Title: "Single task", Description: "Detail"})
```

- **Prefer deterministic tests**: control time, randomness, and database boundaries.
- **Test section headers**: use `// --------------------------------------------------------------------------` to visually separate groups of tests for different endpoints.
- **Test naming**: `Test[FunctionName]_[Scenario]` — e.g., `TestCreateTask_Success`, `TestCreateTask_InvalidBody`, `TestGetTask_NotFound`, `TestUpdateTask_PartialUpdate`, `TestDeleteTask_Success`.
- **Success path tests should**:
  - Assert HTTP status code.
  - Assert key fields in the response body.
  - Assert the object was actually persisted (e.g., for delete, verify fetch returns error).
- **Error path tests should**:
  - Assert the correct HTTP error status.
  - Assert the error response body has a non-empty `"error"` field.

---

## Dependencies & versions

- **Use `go get` with explicit versions**: run `go get pkg@vX.Y.Z` to pin, then `go mod tidy`.
- **Avoid unnecessary dependencies**: prefer Go standard library when reasonable.
- **Current key dependencies**:
  - `github.com/gofiber/fiber/v2` — HTTP framework.
  - `gorm.io/gorm` + `gorm.io/driver/sqlite` — ORM and database drivers.
  - `github.com/go-playground/validator/v10` — request validation.
- **Do not add a new dependency without a clear need** that cannot be met by the standard library or existing deps. If you do add one, document why in the PR/commit.

---

## Reviews & maintenance

- **Leave the codebase better**:
  - Small cleanups are good when they reduce future confusion.
  - Don't mix refactors with feature changes unless necessary.
- **Update docs when behavior changes**:
  - If a change alters how a feature works, update the relevant `docs/` page.
  - If feedback indicates a guideline gap, update this doc so the rule is explicit.

## AI agent rules (explicit)

- **Follow this document by default** when editing or creating files in this repo.
- **Prefer minimal diffs**: don't rewrite unrelated code.
- **Don't introduce new dependencies** without a clear need.
- **If the user corrects you**, treat that correction as a new rule:
  - Update `docs/CODE_GUIDELINES.md` (or a more specific doc) so future agents won't repeat the mistake.
- **Document learned facts**: when you discover something useful about the codebase (a gotcha, a convention, a preference), update `AGENTS.md` or the relevant doc so future agents benefit.
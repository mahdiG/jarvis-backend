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

## Naming

- **Never use 1-2 letter names or unclear abbreviations** — this is a hard rule, not a suggestion. Bad examples: `h`, `b`, `t`, `r`, `hs`, `res`, `obj`, `srv`, `repo`, `svc`, `data`. A name must be meaningful and obvious even outside of its immediate context. If you show the name to a random developer unfamiliar with the code, they should understand what it refers to without explanation.
- **Prefer descriptive names over abbreviations**: use `CreateHabit` instead of `NewH`, `GetHabitByID` instead of `Get`, `IsAuthenticated` instead of `Auth`. No abbreviation is too obvious to skip — always spell out the meaning.
- **The "random person" test**: can you tell a developer unfamiliar with this codebase what a function/variable does just by reading its name? If not, rename it. `GetHabits` passes — it's obvious. `h`, `b`, `r` fail — they carry no meaning to anyone.
- **Prefer domain language**: name things after product concepts (habit, session, streak, event).
- **Boolean naming**:
  - Use `Is/Has/Can/Should` prefixes: `IsAuthenticated`, `HasUnsavedChanges`.
  - Avoid double negatives.
- **Function naming**:
  - Prefer verbs: `CreateHabit`, `GetHabitByID`, `DeleteHabit`, `ParseSchedule`.
  - Function names must be unambiguous. `Get` is never acceptable — use `GetHabitByID`, `GetActiveGoals`, etc. A function name should make the return value obvious without reading its signature.
- **Case conventions** (Go standard):
  - **PascalCase** for: exported functions (`CreateHabit`), exported types (`Habit`, `HabitController`), exported constants (`DefaultPort`, `UIDLength`).
  - **camelCase** for: unexported functions (`parseSchedule`), unexported types (`habitRepository`), variables (`completedCount`), local constants.
  - **snake_case** for: JSON field names (`"created_at"`), database column names (`updated_at`), environment variable names (`DATABASE_DSN`).
  - **kebab-case** for: URL paths (`/api/habits/:id`), middleware names, route group prefixes.
  - Never mix: don't use PascalCase in JSON tags or snake_case in Go identifiers.
- **Directory and file names**:
  - **Directories**: use `kebab-case`. May be singular or plural depending on what reads naturally (e.g., `repositories/`, `controllers/`, `utils/`, `configs/`).
  - **Go source files**: use `kebab-case`. File names are usually singular — they reflect the primary type or responsibility (e.g., `habit.go`, `focus-session.go`, `weekly-review.go`, not `habits.go` or `focus-sessions.go`).
  - **One primary exported type per file**: the file name matches that type. Cohesive helpers for the type may live in the same file. Avoid splitting one type across files or cramming many unrelated types into one file.
- **Package naming**: package names must match their directory name — short, lowercase, no underscores, no mixed casing. For example, `package controllers` lives in `controllers/`, `package router` lives in `router/`. Avoid generic names like `package common` or `package util`.
- **Export discipline**: start with unexported identifiers. Export only when an external package genuinely needs access. This keeps the public surface minimal and intentional. Avoid "export and regret" — unexporting is a breaking change for consumers.
- **Import naming**: rely on Go's default import naming (the last segment of the package path). Use import aliases only to resolve collisions or when the default name is unclear. Never rename imports arbitrarily.

## Functions, modules, and structure

- **Keep functions short**: if a function is hard to scan, split it.
- **One reason to change**: each function should have a single responsibility.
- **Avoid repetition**:
  - Extract reusable helpers and modules.
  - Prefer shared utilities over copy/paste.
  - If logic is reused in 3+ places, it likely belongs in a shared module.
- **Keep packages focused**:
  - `models/` — data structures, type definitions, and validation.
  - `repositories/` — database access and queries (GORM). No business logic.
  - `controllers/` — HTTP handlers (Fiber), request parsing, response formatting. Keep thin.
  - `services/` — business logic (if needed). Testable without HTTP.
  - `configs/` — initialization, wiring, and configuration.
  - `router/` — route definitions and middleware setup.
  - `utils/` — shared helpers (only when logic is truly cross-cutting).
- **Layered dependency direction**: controllers → services → repositories → models. Never reverse.
- **Minimize side effects**: isolate IO (database, network, filesystem) behind repository/service interfaces.

## Go & types

- **Don't fight the type system**: prefer specific types over `interface{}` or `any`. Model domain states precisely.
- **Prefer value types over pointers** for small, immutable data (e.g., `UID` as `string`, `Habit` as value when not mutated). Use pointers for mutation or optional fields.
- **Use custom types for domain primitives**: use named types (e.g., `type UID string`) instead of raw `string` to convey meaning and prevent misuse.
- **Return concrete types from public functions** unless you need abstraction. Introduce interfaces where callers need polymorphism.
- **Zero-value initialization**: prefer zero values over `nil` where possible (e.g., `var h Habit` instead of `nil`).
- **Avoid `panic`** in application code. Only use in truly unrecoverable situations (startup init).
- **Error handling**:
  - Return errors, don't ignore them. Respect `_ = fn()` only when intentionally discarding a result.
  - Use `fmt.Errorf("context: %w", err)` to wrap errors with context.
  - Use `errors.Is` / `errors.As` for sentinel/wrapped error checking.
  - Define sentinel errors (e.g., `var ErrNotFound = errors.New("not found")`) for expected failure cases.
- **Runtime validation for external data**: anything coming from network (HTTP requests) or storage (DB) is untrusted — validate/parse before use.

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
              "error": "habit not found",
          })
      }
      slog.Error("failed to get habit", "id", id, "error", err)
      return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
          "error": "failed to get habit from database",
      })
  }
  ```

## HTTP handler code (Fiber)

- **Handlers should be thin**: parse request, delegate to service/repository, format response.
- **Use consistent JSON response format**: `{"error": "..."}` for errors, direct object/array for success.
- **Use Fiber's `c.Params`, `c.Query`, `c.BodyParser`** for input extraction.
- **Validate parsed input** before passing to business logic.
- **Use `fiber.Map`** for ad-hoc JSON maps, or dedicated response types for complex structures.
- **Use GORM model structs as API contracts directly** — do not create separate request or response DTOs. The `models/` package owns both the database schema and the API contract. Add JSON tags (snake_case) and validation tags directly to the model struct. The frontend sends/receives JSON that maps to these same struct fields. This keeps the API surface in sync with the data layer and eliminates translation code.


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
- **Import ordering**: standard library → third-party → internal modules, separated by blank lines.
- **Comments**:
  - Prefer explaining *why* a decision was made, not *what* the code does (the code is the *what*).
  - Every exported identifier must have a doc comment (`// PackageName` for packages, `// FunctionName` for functions/types).
  - Unexported helpers that are non-obvious should also have a comment.
  - Keep comments up to date — a stale comment is worse than no comment.
- **Receiver naming**: use meaningful receiver names that match the type. Prefer descriptive abbreviations over single letters (e.g., `habit Habit`, `service HabitService`). Be consistent across all methods on the same type.

## Dependencies & versions

- **Use `go get` with explicit versions**: run `go get pkg@vX.Y.Z` to pin, then `go mod tidy`.
- **Avoid unnecessary dependencies**: prefer Go standard library when reasonable.

## Testing (required for non-trivial changes)

Tests are how we keep the app stable and make refactors safe.

- **Write tests for**:
  - Bug fixes (regression tests).
  - Non-trivial business logic (parsing, calculations, state transitions).
  - Repository/controller integration flows (when feasible).
- **Use the right test approach**:
  - `testing` package for all tests.
  - Use `testify/assert` or `testify/require` for readable assertions.
  - Prefer table-driven tests for multiple cases.
- **Prefer deterministic tests**: control time, randomness, and database boundaries.
  - Use in-memory SQLite or testcontainers for repository tests.
  - Mock external dependencies where appropriate.
- **Test behavior, not implementation**: assert outcomes and API contract, not internal calls.
- **Keep tests readable**: arrange/act/assert structure, descriptive test names (`TestGetHabit_WhenNotFound`).
- **Test file location**: place `_test.go` files in the same package as the code they test (white-box) or in an `_test` package (black-box) as appropriate.
- **Test data**: use small, focused test helpers to build fixtures. Avoid large JSON files or shared global state across tests. Each test should set up its own data.

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
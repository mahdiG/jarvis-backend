# Agent Tool Functions — Adding & Guidelines

This document describes how LLM-accessible tool functions are defined, implemented, and registered in the Jarvis backend. It is the **source of truth** for anyone (human or AI) adding or modifying agent tools.

> **Related files**: `agent/tool.go`, `agent/agent.go`, `agent/types.go`

---

## Table of Contents

1. [Overview](#overview)
2. [File Layout (agent/tool.go)](#file-layout-agenttoolgo)
3. [Step-by-Step: Adding a New Tool](#step-by-step-adding-a-new-tool)
4. [Naming Conventions (Non‑Negotiable)](#naming-conventions-non-negotiable)
5. [Schema Generation Patterns](#schema-generation-patterns)
6. [Executor Function Patterns](#executor-function-patterns)
7. [Connecting to the Agent](#connecting-to-the-agent)
8. [Existing Tools Reference](#existing-tools-reference)

---

## Overview

Agent tool functions are Go functions in the `agent` package that the LLM can call during a chat session. Each tool has:

- A **name constant** — the string the LLM uses to request the tool (e.g. `"create_task"`).
- A **`getToolInfo*` function** — returns a `*schema.ToolInfo` that defines the tool's name, description, and JSON Schema for its parameters.
- An **`executeTool*` function** — the actual handler that receives the LLM's JSON arguments, performs the operation, and returns a JSON result string.

Tools are registered in `agent.go:Init()` and dispatched in `agent.go:executeTool()`.

---

## File Layout (`agent/tool.go`)

The file is organized in sections, top to bottom:

1. **Tool name constants** — grouped in a single `const` block.
2. **Schema helper functions** — `getToolSchemaFromModel` and `getBaseFieldPropertyNames`.
3. **Per-tool `getToolInfo*` functions** — one per tool, ordered alphabetically by tool name.
4. **Per-tool `executeTool*` functions** — one per tool, ordered alphabetically to match the info functions.

```
// ── Constants ──────────────────────────────────────────────
const (
    toolNameCreateTask = "create_task"
    toolNameGetTasks   = "get_tasks"
    …
)

// ── Schema helpers ─────────────────────────────────────────
func getToolSchemaFromModel[Model any](shouldExcludeBase bool) *schema.ParamsOneOf { … }
func getBaseFieldPropertyNames() []string { … }

// ── Tool info functions ────────────────────────────────────
func getToolInfoCreateTask() *schema.ToolInfo { … }
func getToolInfoGetTasks()   *schema.ToolInfo { … }
…

// ── Tool executor functions ────────────────────────────────
func executeToolCreateTask(_ context.Context, argumentsJSON string) (string, error) { … }
func executeToolGetTasks(_ context.Context, argumentsJSON string) (string, error) { … }
…
```

---

## Step-by-Step: Adding a New Tool

### Step 1 — Define the tool name constant

Add a new constant to the `const` block in `agent/tool.go`:

```go
const (
    // … existing constants …
    toolNameDeleteHabit = "delete_habit"
)
```

The constant name must follow the `toolNameVerbNoun` pattern (see [Naming Conventions](#naming-conventions-non-negotiable)).

### Step 2 — Create the `getToolInfo*` function

Choose the schema generation approach that fits the tool:

**Option A — Model-based (most common for CRUD on a model):**

```go
func getToolInfoCreateHabit() *schema.ToolInfo {
    return &schema.ToolInfo{
        Name:        toolNameCreateHabit,
        Desc:        "Create a new habit. Use this when the user asks to create, add, or start a new habit",
        ParamsOneOf: getToolSchemaFromModel[models.Habit](true),
    }
}
```

Pass `true` for `shouldExcludeBase` when the LLM should not provide ID/timestamps (e.g. for create). Pass `false` when the LLM needs to specify an ID (e.g. get/update/delete by ID).

**Option B — Manual params (for tools with custom arguments like pagination):**

```go
func getToolInfoGetHabits() *schema.ToolInfo {
    return &schema.ToolInfo{
        Name: toolNameGetHabits,
        Desc: "Get habits. Use limit=0 to return all habits, or limit=N to return the first N habits.",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "limit": {
                Desc:     "Maximum number of habits to return. 0 returns all.",
                Type:     schema.Integer,
                Required: true,
            },
            "offset": {
                Desc: "Number of habits to skip before returning results.",
                Type: schema.Integer,
            },
        }),
    }
}
```

### Step 3 — Create the `executeTool*` function

Write the handler following the [Executor Function Patterns](#executor-function-patterns):

```go
func executeToolCreateHabit(_ context.Context, argumentsJSON string) (string, error) {
    var habit models.Habit
    err := json.Unmarshal([]byte(argumentsJSON), &habit)
    if err != nil {
        return "", utils.WrapError(err)
    }

    // Validate required inputs (the model's validate tag may handle this,
    // but always do a fast check for required primitives).
    if habit.Name == "" {
        return "", utils.WrapError(errors.New("habit name is required"))
    }

    created, err := repositories.CreateHabit(habit)
    if err != nil {
        return "", utils.WrapError(err)
    }

    result, marshalError := json.Marshal(created)
    if marshalError != nil {
        return "", utils.WrapError(marshalError)
    }
    return string(result), nil
}
```

### Step 4 — Register the tool in `agent.go:Init()`

Add the `*schema.ToolInfo` to the slice passed to `einoagent.ChatModelWithTools`:

```go
chatModelWithTools, err := einoagent.ChatModelWithTools(
    openaiCompatibleChatModel,
    openaiCompatibleChatModel,
    []*schema.ToolInfo{
        getToolInfoCreateTask(),
        getToolInfoGetTasks(),
        // … add the new one here:
        getToolInfoCreateHabit(),
        getToolInfoGetHabits(),
    },
)
```

Maintain alphabetical order by tool name in the slice.

### Step 5 — Add the dispatch case in `agent.go:executeTool()`

Add a new `case` to the switch statement:

```go
func executeTool(ctx context.Context, toolCall schema.ToolCall) (string, error) {
    switch toolCall.Function.Name {
    case toolNameCreateTask:
        return executeToolCreateTask(ctx, toolCall.Function.Arguments)
    case toolNameGetTasks:
        return executeToolGetTasks(ctx, toolCall.Function.Arguments)
    // … add the new one here:
    case toolNameCreateHabit:
        return executeToolCreateHabit(ctx, toolCall.Function.Arguments)
    case toolNameGetHabits:
        return executeToolGetHabits(ctx, toolCall.Function.Arguments)
    default:
        return "", utils.WrapError(errors.New("unknown tool: " + toolCall.Function.Name))
    }
}
```

Maintain alphabetical order by `case` label.

---

## Naming Conventions (Non‑Negotiable)

These rules are **mandatory**. A PR that violates them will be rejected.

| Element | Pattern | Example |
|---|---|---|
| Tool name constant | `toolNameVerbNoun` | `toolNameCreateTask`, `toolNameGetTask` |
| Tool name string (LLM-facing) | `snake_case`, verb_noun | `"create_task"`, `"get_task"` |
| Info function | `getToolInfoVerbNoun` | `getToolInfoCreateTask` |
| Executor function | `executeToolVerbNoun` | `executeToolCreateTask` |

### Description style

- Use **imperative tone**: "Create a new task…", "Delete a task by its ID…".
- Include **example trigger phrases** for the LLM: "Use this when the user asks to create, add, or make a task."
- Keep descriptions concise but actionable. The LLM reads them to decide which tool to call.

### Prohibited

- **Never use 1-2 letter names or unclear abbreviations** in any identifier (constant name, function name, parameter name, variable name inside the executor).
  - Bad: `toolNameCT`, `getInfoHT`, `executeHT`, `h`, `r`, `t`, `s`, `hs`, `svc`.
  - Good: `toolNameCreateHabit`, `getToolInfoCreateHabit`, `executeToolCreateHabit`.
- **Never skip the `toolName` or `getToolInfo` or `executeTool` prefix** — these prefixes prevent collisions and make the code self-documenting.
- **Never use tool names that don't start with a verb** — every tool must perform an action (`create_task`, not `task_creator`).

---

## Schema Generation Patterns

### `getToolSchemaFromModel[T](shouldExcludeBase bool)`

This generic helper generates a JSON Schema for a tool's parameters by reflecting the fields of a model struct.

- **`shouldExcludeBase = true`**: Strips fields from `models.Base` (ID, CreatedAt, UpdatedAt, DeletedAt) from the schema. Use for create operations where the LLM should not provide auto-generated fields.
- **`shouldExcludeBase = false`**: Exposes all fields including ID. Use for get/update/delete by ID operations where the LLM must specify which record to act on.

**Important**: The reflector uses `ExpandedStruct: true` to inline the schema definition so OpenAI-compatible providers accept it. Do not remove this flag.

### Manual params (`schema.NewParamsOneOfByParams`)

Use for tools that take custom arguments not directly corresponding to a model struct — for example, pagination (`limit`, `offset`), search queries, or configuration.

### Choosing between them

| Situation | Approach |
|---|---|
| CRUD on a model entity | `getToolSchemaFromModel` |
| Paginated listing | `getToolSchemaFromModel` on the model with manual `limit`/`offset` overrides, OR full manual params |
| Utility/action that doesn't map to a model | Manual `NewParamsOneOfByParams` |
| Tool where you need to exclude specific fields (not just Base) | Manual `NewParamsOneOfByParams` |

---

## Executor Function Patterns

### Signature

Every executor must have this exact signature:

```go
func executeToolVerbNoun(_ context.Context, argumentsJSON string) (string, error)
```

- The `context.Context` parameter is reserved for future use (tracing, cancellation). The name is `_` to signal it is intentionally unused.
- `argumentsJSON` is the raw JSON string the LLM provided for the tool call.
- The return is always `(string, error)` — a JSON-encoded result string, or an error.

### Standard body pattern

```
1. Parse JSON arguments into a typed struct.
2. Validate required fields (fast guard checks).
3. Execute the repository/service call.
4. Marshal the result to JSON.
5. Return the JSON string (or error).
```

```go
func executeToolVerbNoun(_ context.Context, argumentsJSON string) (string, error) {
    // 1. Parse
    var input models.SomeType
    err := json.Unmarshal([]byte(argumentsJSON), &input)
    if err != nil {
        return "", utils.WrapError(err)
    }

    // 2. Validate
    if input.SomeRequiredField == "" {
        return "", utils.WrapError(errors.New("some field is required"))
    }

    // 3. Execute
    result, err := repositories.SomeOperation(input)
    if err != nil {
        return "", utils.WrapError(err)
    }

    // 4. Marshal
    resultJSON, marshalError := json.Marshal(result)
    if marshalError != nil {
        return "", utils.WrapError(marshalError)
    }

    // 5. Return
    return string(resultJSON), nil
}
```

### Error handling rules

- **Always use `utils.WrapError`** when creating or returning errors. This captures the source file:line in the log.
- **Return validation errors** as `"", utils.WrapError(errors.New("human readable message"))`.
- **Return repository errors** as `"", utils.WrapError(err)`.
- **Do not panic**. Return errors instead.
- **The agent loop in `agent.go:Chat()`** already catches errors from executors and feeds them back to the LLM as tool result messages, so the LLM can self-correct or explain the failure to the user.

### Return value rules

- **Success**: return the JSON-encoded result as a string (e.g. `string(resultJSON), nil`).
- **Delete operations**: return a plain success message (e.g. `"task deleted successfully", nil`) rather than JSON.
- **Single-entity operations**: return the JSON of the created/updated/retrieved entity.
- **List operations**: return a JSON array of entities (even if empty — `"[]"`).

---

## Connecting to the Agent

Two sites in `agent/agent.go` must be updated when adding a tool:

**1. Registration in `Init()`** — the tool is added to the `[]*schema.ToolInfo` slice so the LLM knows it exists:

```go
// agent/agent.go — Init()
chatModelWithTools, err := einoagent.ChatModelWithTools(
    openaiCompatibleChatModel,
    openaiCompatibleChatModel,
    []*schema.ToolInfo{
        getToolInfoCreateTask(),
        // … add new getToolInfo*() call here
    },
)
```

**2. Dispatch in `executeTool()`** — the tool name constant is mapped to the executor function:

```go
// agent/agent.go — executeTool()
switch toolCall.Function.Name {
case toolNameCreateTask:
    return executeToolCreateTask(ctx, toolCall.Function.Arguments)
// … add new case here
default:
    return "", utils.WrapError(errors.New("unknown tool: " + toolCall.Function.Name))
}
```

Both lists must be kept in **alphabetical order by tool name** for maintainability.

---

## Existing Tools Reference

| Tool Name | Constant | Info Function | Executor Function | Schema Type |
|---|---|---|---|---|
| `create_task` | `toolNameCreateTask` | `getToolInfoCreateTask` | `executeToolCreateTask` | Model (`Task`, exclude Base) |
| `delete_task` | `toolNameDeleteTask` | `getToolInfoDeleteTask` | `executeToolDeleteTask` | Model (`Task`, include Base) |
| `get_task` | `toolNameGetTask` | `getToolInfoGetTask` | `executeToolGetTask` | Model (`Task`, include Base) |
| `get_tasks` | `toolNameGetTasks` | `getToolInfoGetTasks` | `executeToolGetTasks` | Manual (`limit`, `offset`) |
| `update_task` | `toolNameUpdateTask` | `getToolInfoUpdateTask` | `executeToolUpdateTask` | Model (`Task`, include Base) |

---

## Checklist: Adding a New Tool

- [ ] 1. Add `toolNameVerbNoun` constant in `agent/tool.go`'s const block
- [ ] 2. Write `getToolInfoVerbNoun()` with the right schema approach
- [ ] 3. Write `executeToolVerbNoun()` following the executor patterns
- [ ] 4. Add `getToolInfoVerbNoun()` to the `Init()` registration slice in `agent.go`
- [ ] 5. Add `case toolNameVerbNoun:` to the `executeTool()` switch in `agent.go`
- [ ] 6. Verify alphabetical ordering in both registration and dispatch
- [ ] 7. Add tests for the new tool's executor function
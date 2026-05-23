2026

May 6 - 16:06
Wrote the jarvis vision with cursor.

May 6 - 21:00
Created frontend project template

May 7 - 15:30
Got custom api key working in vscode with Continue extension and created test login and habits pages.
Created jarvis design system, timeline page with Cursor

May 7 - 19:22
Research Lit-labs router. Write code guidelines doc with cursor

May 7 - 19:52
create AGENTS.md

May 7 - 19:58
Update cursor shortcuts

May 7 - 20:11
Create app router

May 8 - 12:00
Buy deepseek api, setup Continue (sucks). Try windsurf, cursor custom api key(requires subscription). Copilot with custom api key (failed). Antigravity (failed to verify account). Cline vscode extension works mostly fine.

May 8 - 03:00 - 09:00
Add localization, make event timeline page functional. AI struggle a little with it but works. Add CRUD to habits.

May 8 - 00:00
add settings store (lit signal). Also set/get locale to/from localstorage

May 9 - 18:00
Setup golang backend project structure

May 10 - 13:02
Think about database schema.
What I want: Value/Wish/Identity (Musician) -> Goals (roadmaps) -> tasks
But I don't want to design 10 layers of subtasks. Maybe with a flat/infinite system I can just use infinite subtasks:

type Task struct {
ID uint `gorm:"primaryKey"`
Title string
Description string
ParentID *uint
Status string // active, done, archived
Type string // "value", "goal", "roadmap", "task", "habit"
ScheduledFor *time.Time
Metadata datatypes.JSON `gorm:"type:json"` // e.g., {"ai_breakdown": [...], "score": 0.7}
CreatedAt time.Time
UpdatedAt time.Time
}

AI SAYS: And a content note
This exact decision—"why I used a single recursive table instead of premature micro-categorization for my AI Life OS"—is the kind of piece that resonates deeply with senior engineers. It shows you know how to balance immediate delivery with long-term flexibility, and that you're not afraid to start simple. Write it down now, even as a paragraph in your progress log.

Fighting temptations to use new shiny technologies and stay with boring mature tech that works and ships fast.

May 11 - 09:20
Implement AI chat API using CloudWeGo Eino library. Added:

- Agent service (services/agent.go) using OpenAI-compatible LLM with tool calling
- Chat controller (controllers/chat.go) with POST /v1/chat endpoint
- create_task tool so the AI can create tasks on behalf of the user
- Environment variables: LLM_API_KEY, LLM_MODEL, LLM_BASE_URL, LLM_MAX_TOKENS, LLM_TEMPERATURE

May 11 - 09:53
Added centralized config with Viper:

- Created configs/env.go with typed Config struct and LoadConfig()
- Broke DATABASE_DSN into individual DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
- Created .env.example documenting all env vars
- Updated services/agent.go to accept typed LLM config struct
- Updated configs/database.go to build DSN from config fields
- Updated cmd/api/main.go to use configs.LoadConfig()
- .env already in .gitignore

May 11 - 13:37
Fixed "Invalid schema for function 'create_task': schema must be a JSON Schema of 'type: \"object\"', got 'type: null'" error.
Root cause: `jsonschema.Reflector` returned a `$ref`-wrapped schema (pointing to a $defs entry) instead of an inline object, so the root-level `type` field was absent — causing OpenAI to reject it with `type: null`.
Fix: Added `ExpandedStruct: true` to the reflector in `getToolSchemaFromModel()` so it expands the root struct and produces a schema with `type: "object"`.
Also improved error tracing: wrapped `chatModel.Generate` errors with `fmt.Errorf` so agent errors show the specific call site in logs.

May 11 - 14:49
Fixed multi-turn tool calling in agent/Chat(): the previous code only supported one round of tool execution — it processed the first batch of tool calls, made a single follow-up LLM call, and only used its text content, discarding any follow-up tool calls (e.g. creating a child task after a parent). Fix: wrapped the generate→process cycle in a `for` loop that continues feeding tool results back until the LLM responds with text only.

May 13 - 09:00
Added Conversations & Messages API for a ChatGPT-like chat UI:

- Models: Conversation (title) and Message (conversationId, role, content)
- Repositories with full CRUD for conversations, create & list for messages
- Controllers under /v1/conversations with nested /messages routes
- SendMessage endpoint integrates with the AI agent, returns both user + assistant messages
- Old controllers/chat.go removed (replaced by conversation flow)
- OpenAPI spec at api/openapi.json, served at GET /v1/swagger/doc.json
- 22 tests pass (14 conversation/message + 8 existing task)
  </replace_in_file>

May 13 - 10:00
Replace openAPI with swaggo (has dashboard). Upgrade fiber to version 3 (I thought we were on the latest version already!)

May 19 - 13:00
Create unified api response type

May 23 - 17:18
Added trash APIs for notes (GetTrashNotes, RestoreNote, PermanentDeleteNote) and documented gotchas:

Key lessons learned (documented in AGENTS.md and docs/ADDING-APIS.md):

1. **Route ordering (Fiber)**: Static sub-paths like `/trash` must be registered BEFORE parameterized paths like `/:id` — otherwise Fiber matches `trash` as the `:id` parameter.
2. **Soft-delete GORM patterns**: Documented how to query trashed items (`Unscoped().Where("deleted_at IS NOT NULL")`), restore (`Update("deleted_at", nil)`), and permanently delete (`Unscoped().Delete()`).
3. **Test helper gotchas**: `DecodeResponseData` needs a concrete pointer (not nil). JSON null check in tests must use `string(env.Data) != "null"` not `env.Data != nil` since `json.RawMessage` from null is a non-nil byte slice.
4. **GORM hard-delete**: `Unscoped().Delete()` with a model that has `gorm.DeletedAt` actually removes the row from the database.

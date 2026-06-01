# ⚠️ Work in Progress

> This project is under **active development**. APIs are unstable and may break without notice. **Not recommended for production use.**

---

# Jarvis — AI Life OS Backend

**Jarvis** is the backend for a personal **AI Life Operating System** — a private, audit-able platform that helps you become a 10x human by capturing life data, turning it into clear decisions and next actions, and eventually taking approved actions in the real world.

This repository contains the Go API server with an LLM-powered agent that understands natural language and can manage tasks, notes, and conversations.

## Features

- **Task management** — CRUD with scheduling (`ScheduledFrom`/`ScheduledTo`), scoring (`TargetScore`/`Score`), subtasks via `ParentID`, and tagging
- **Note management** — Rich notes with content, tags, and file attachments
- **Conversations & chat** — Persistent conversations with an AI agent that can create, read, update, and delete tasks via natural language
- **AI agent** — Powered by [Eino](https://github.com/cloudwego/eino) (CloudWeGo's LLM framework) with OpenAI-compatible providers (OpenAI, DeepSeek, Ollama, etc.)
- **Soft-delete & trash** — All entities support soft-delete with restore and permanent deletion
- **Swagger / OpenAPI** — Auto-generated API docs at `/v1/swagger/`
- **PostgreSQL** — Production database via Docker Compose
- **Tagging system** — Flexible many-to-many tagging across tasks and notes

## Tech Stack

| Layer      | Technology                                                              |
| ---------- | ----------------------------------------------------------------------- |
| Language   | Go 1.26                                                                 |
| HTTP       | [Fiber v3](https://github.com/gofiber/fiber)                            |
| Database   | PostgreSQL 18 (production), SQLite (dev/testing)                        |
| ORM        | [GORM](https://gorm.io)                                                 |
| AI Agent   | [Eino](https://github.com/cloudwego/eino) + OpenAI-compatible providers |
| Validation | [go-playground/validator](https://github.com/go-playground/validator)   |
| Config     | [Viper](https://github.com/spf13/viper)                                 |
| API Docs   | [swaggo/swag](https://github.com/swaggo/swag)                           |

## Quick Start

### Prerequisites

- Go 1.26+
- Docker & Docker Compose (for PostgreSQL)

### 1. Clone and configure

```bash
git clone https://github.com/mahdiG/jarvis-backend.git
cd jarvis-backend
cp .env.example .env
```

Edit `.env` to set your `LLM_API_KEY` (required) and adjust other settings as needed.

### 2. Start PostgreSQL

```bash
docker compose up -d
```

### 3. Run the server

```bash
go run ./cmd/api
```

The server starts on `http://localhost:3000` by default.

### 4. Try the AI chat

```bash
curl -X POST http://localhost:3000/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"title": "My first chat"}'

# Copy the conversation ID from the response, then:
curl -X POST http://localhost:3000/v1/conversations/<ID>/messages \
  -H "Content-Type: application/json" \
  -d '{"role": "user", "content": "Create a task called Buy groceries with description Remember to buy milk and eggs"}'
```

### 5. Browse the API docs

Open `http://localhost:3000/v1/swagger/` in your browser.

## Configuration

Configuration is loaded from a `.env` file (or OS environment variables). See [.env.example](.env.example) for all available options.

| Variable               | Default       | Description                                |
| ---------------------- | ------------- | ------------------------------------------ |
| `SERVER_PORT`          | `3000`        | HTTP server port                           |
| `CORS_ALLOWED_ORIGINS` | `*`           | CORS allowed origins                       |
| `DB_HOST`              | `localhost`   | PostgreSQL host                            |
| `DB_PORT`              | `5432`        | PostgreSQL port                            |
| `DB_USER`              | `jarvis`      | PostgreSQL user                            |
| `DB_PASSWORD`          | `jarvis-pass` | PostgreSQL password                        |
| `DB_NAME`              | `jarvis`      | PostgreSQL database name                   |
| `DB_SSLMODE`           | `disable`     | PostgreSQL SSL mode                        |
| `LLM_API_KEY`          | —             | **Required.** API key for the LLM provider |
| `LLM_MODEL`            | `gpt-4o`      | Model name                                 |
| `LLM_BASE_URL`         | —             | Base URL for non-OpenAI providers          |
| `LLM_MAX_TOKENS`       | `4096`        | Max completion tokens                      |
| `LLM_TEMPERATURE`      | `0.7`         | Model temperature                          |

## API Overview

All endpoints are prefixed with `/v1`.

### Tasks

| Method   | Path           | Description             |
| -------- | -------------- | ----------------------- |
| `GET`    | `/tasks`       | List tasks              |
| `POST`   | `/tasks`       | Create tasks            |
| `PATCH`  | `/tasks`       | Update tasks            |
| `DELETE` | `/tasks`       | Soft-delete tasks       |
| `GET`    | `/tasks/trash` | List soft-deleted tasks |
| `PATCH`  | `/tasks/trash` | Restore tasks           |
| `DELETE` | `/tasks/trash` | Permanently delete      |
| `GET`    | `/tasks/:id`   | Get task by ID          |

### Notes

Same pattern as tasks — CRUD + trash management.

### Tags

| Method   | Path        | Description   |
| -------- | ----------- | ------------- |
| `GET`    | `/tags`     | List tags     |
| `POST`   | `/tags`     | Create tags   |
| `PATCH`  | `/tags`     | Update tags   |
| `DELETE` | `/tags`     | Delete tags   |
| `GET`    | `/tags/:id` | Get tag by ID |

### Conversations & Messages

| Method   | Path                          | Description           |
| -------- | ----------------------------- | --------------------- |
| `GET`    | `/conversations`              | List conversations    |
| `POST`   | `/conversations`              | Create conversation   |
| `GET`    | `/conversations/:id`          | Get conversation      |
| `PATCH`  | `/conversations/:id`          | Update conversation   |
| `DELETE` | `/conversations/:id`          | Delete conversation   |
| `GET`    | `/conversations/:id/messages` | List messages         |
| `POST`   | `/conversations/:id/messages` | Send message to agent |

Full OpenAPI documentation is available at `/v1/swagger/doc.json` and browsable at `/v1/swagger/`.

## Project Structure

```
├── agent/              # AI agent — tool definitions, LLM chat loop
├── cmd/api/            # Application entry point (main.go)
├── configs/            # Configuration loading and database init
├── constants/          # Shared constants
├── controllers/        # HTTP handlers (Fiber)
├── docs/               # Documentation
├── models/             # Data structures and GORM models
├── repositories/       # Database access layer
├── router/             # Route definitions
└── utils/              # Shared utilities (logging, error wrapping)
```

## Development

### Run tests

```bash
go test ./...
```

### Regenerate Swagger docs

```bash
swag init --output docs/swagger --generalInfo controllers/doc.go --parseDependency
```

### Architecture

The codebase follows a layered architecture with clean dependency direction:

```
controllers → repositories → models
```

- **Controllers** are thin HTTP handlers that parse requests and format responses
- **Repositories** handle database operations with no business logic
- **Models** define data structures, validation rules, and GORM hooks

See [docs/CODE-GUIDELINES.md](docs/CODE-GUIDELINES.md) for detailed coding standards.

## Vision

Jarvis aims to become a **self-quantified life OS** — a private personal growth system that:

1. **Captures** daily data (tasks, habits, mood, notes, schedule)
2. **Reflects** through weekly reviews and trend analysis
3. **Decides** on high-leverage next actions
4. **Executes** with reminders and (eventually) approved automations
5. **Learns** your patterns and adapts over time

Read the full vision at [docs/JARVIS-LIFE-OS-VISION.md](docs/JARVIS-LIFE-OS-VISION.md).

## Documentation

- [Code Guidelines](docs/CODE-GUIDELINES.md) — Coding standards and conventions
- [Adding New APIs](docs/ADDING-APIS.md) — Step-by-step guide for adding endpoints
- [Agent Tool Functions](docs/AGENT-TOOL-FUNCTIONS.md) — How the AI agent works
- [Life OS Vision](docs/JARVIS-LIFE-OS-VISION.md) — Product vision and roadmap

## License

MIT

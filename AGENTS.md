# AI Agents Guide (Jarvis Backend)

This repository is worked on by humans and AI agents (Cursor and others). This file tells agents how to operate here.

Agents should use their tool functions when needed and prefer it over running shell commands if possible.

## Source of truth docs (read first)

- **Code guidelines (mandatory)**: `docs/CODE_GUIDELINES.md`
- **Vision / product direction**: `docs/JARVIS_LIFE_OS_VISION.md`
- **Progress / current focus**: `docs/progress.md`

If there is any conflict, follow **`docs/CODE_GUIDELINES.md`** for coding standards and ask for clarification only if requirements are truly ambiguous.

## Non-negotiable agent rules

- **Follow `docs/CODE_GUIDELINES.md`** when creating or editing any code.
- **Prefer clarity over cleverness** and use complete, understandable names.
- **Never use 1-2 letter names or unclear abbreviations** for variables, functions, types, or receivers. A name must be meaningful and obvious even outside its immediate context. `GetHabits` is clear — `h`, `b`, `r`, `hs`, `svc`, `repo` are not. Always spell out the meaning.
- **Keep changes minimal and focused**: don't refactor unrelated code while implementing a feature/fix.
- **Avoid repetition**: extract reusable functions/packages when logic is repeated.
- **Write tests for non-trivial changes** (especially bug fixes and core logic), per `docs/CODE_GUIDELINES.md`.

## Updating docs from feedback (required)

If the user gives feedback like:

- "Don't do X again"
- "Prefer Y style/pattern"
- "This project should always do Z"

…then the agent must **update the relevant documentation** so the rule is captured and future work doesn't repeat the mistake.

Default behavior:

- Update `docs/CODE_GUIDELINES.md` for coding/process rules.
- Update `docs/JARVIS_LIFE_OS_VISION.md` for product principles and direction.

## Documenting learned facts (required)

Whenever an AI agent learns a useful fact about the codebase, a user preference, a project convention, a gotcha, or anything else that would make it easier for future agents (or humans) to work effectively here, it must immediately document it in the appropriate file.

Types of facts to document:

- **Codebase insights**: non-obvious structure, tricky dependencies, architectural decisions, known quirks.
- **User preferences**: naming conventions, style choices, tooling preferences, process preferences.
- **Project gotchas**: common pitfalls, tests that fail for subtle reasons, environment setup quirks.
- **Workflow improvements**: scripts, aliases, commands, or patterns that save time.

Default behavior:

- Update `AGENTS.md` for agent-specific rules, workflow tips, and cross-cutting conventions.
- Update `docs/CODE_GUIDELINES.md` for coding standards, patterns, or style rules.
- Update `docs/progress.md` for notable changes in project state or direction.

When in doubt, add the fact to `AGENTS.md` rather than leaving it undocumented. A short note beats tribal knowledge.

## Working agreements

- **No dependency changes** unless clearly necessary and aligned with the repo direction.
- **Keep errors actionable**: no silent failures; return meaningful error messages and structured HTTP responses.
- **Match existing conventions** in the touched area unless improving consistency is explicitly requested.
- **Layered architecture**: controllers → services → repositories → models. Keep the dependency direction clean.
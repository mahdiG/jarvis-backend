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

## Working agreements

- **No dependency changes** unless clearly necessary and aligned with the repo direction.
- **Keep errors actionable**: no silent failures; return meaningful error messages and structured HTTP responses.
- **Match existing conventions** in the touched area unless improving consistency is explicitly requested.
- **Layered architecture**: controllers → services → repositories → models. Keep the dependency direction clean.
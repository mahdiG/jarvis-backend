# Jarvis Life OS — Vision, Principles, and Feature Spec

## North Star
Build a **Jarvis-like AI Life Operating System** that helps you become a **10x human**: happier, healthier, more productive, and more successful—by **capturing life data**, turning it into **clear decisions and next actions**, and eventually **taking approved actions** in the real world.

The product is meant to run on **Android and Linux** as an app and/or launcher, and to become your daily “home base” for:
- habits
- tasks/projects
- time and schedules
- journaling and mood/energy
- finances
- life reviews and planning

Over time it learns your patterns, builds a personality/work-style model, and uses **scientific behavior-change techniques** to keep you on track—first with suggestions, later with enforcement/guardrails you explicitly enable.

## One-line product definition
**A private, audit-able personal growth OS that turns your daily logs into weekly strategy and approved automation.**

## Who it’s for
### Primary user
- You (single-user first)

### Later users (optional)
- Individuals who want a privacy-respecting personal growth system

## Core promises
- **Fast capture**: logging common things should take seconds.
- **High leverage**: the system should produce decisions, not just dashboards.
- **Capability-first (with safety)**: optimize for a powerful assistant, while keeping strong auditability and user control.
- **Human-in-control**: the assistant never takes external actions without explicit approval (and later, user-configured policies).
- **Truth over vibes**: coaching is grounded in evidence-based techniques, with clear uncertainty and trade-offs.

## What “10x human” means (operationalized)
Jarvis Life OS maximizes:
- **Clarity**: you always know what matters today.
- **Consistency**: habits and routines happen with minimal willpower.
- **Progress**: weekly trajectory improves, not just daily busyness.
- **Energy**: sleep/stress/health are treated as first-class constraints.
- **Focus**: deep work is protected; distractions are constrained.
- **Agency**: you spend less time coordinating life; more time living it.

## Self-quantified life (long-term)
The system should evolve into a **self-quantified life OS**: a rich, structured timeline of what you did, how you felt, and what it cost—so coaching can be personalized and high-confidence.
- the assistant becomes more accurate as data density increases
- analysis moves from generic advice → *your* leading indicators and leverage points
- the product becomes an “exoskeleton for self-improvement”: it notices patterns you can’t and helps you act on them

## Product principles
- **Local-first capture** (especially on Android): capture works offline; sync later.
- **Transparency**: show the “why” behind suggestions; show inputs used.
- **Minimalism in the moment**: “Today” view is simple; detail is available when needed.
- **Compounding system**: every week improves the system’s model and your routines.
- **Safety rails for autonomy**: tight scopes, least privilege, approvals, and audit trails.
- **Adaptive coaching**: tone and interventions adapt to your energy, context, and what historically worked for you.

---

## The Life OS loop (how it works)
Jarvis is not “a chat bot”; it’s a loop:

1) **Capture** (inputs)
- quick logs: habit done, mood, journal, expense, time block, event
- passive imports (later): calendar, sleep/steps, bank transactions, app usage, window/app activity (ActivityWatch-like)

2) **Organize**
- normalize data into consistent categories (work/health/social/finance)
- link items to goals/projects
- detect anomalies and missing data (gentle reminders)

3) **Reflect** (analysis)
- daily recap (optional)
- weekly review (core)
- monthly/quarterly retros
- correlations and leading indicators (e.g. sleep → focus; meetings → stress)

4) **Decide**
- pick 1–3 “big moves” for the week
- create next actions, schedule blocks, adjust habit targets

5) **Execute**
- reminders, checklists, focus mode, timeboxing
- later: integrations and automation (“hands”) with approvals

6) **Learn**
- update your personal model, preferences, and effective interventions
- propose memory updates (you approve)

---

## Core modules (feature areas)

### 1) Today (command center)
**Goal**: a single screen that tells you what matters now.
- daily priorities (Top 3)
- schedule overview + next block
- habit checklist (only those due today)
- “quick capture” actions (log mood, journal, expense, timer)
- warnings (overbooked day, low sleep, too many tasks, spending spike)

### 1.1) Schedule (timeline-first)
**Goal**: your day as a live timeline showing where you are and what to do next.
- “Now” indicator on a vertical timeline
- current block + next block + upcoming transitions
- quick actions: start/stop, skip, reschedule, shorten/extend, attach notes
- timeboxing support: tasks become blocks; blocks link back to tasks/goals
- later: proactive nudges (“it’s time to do X”) and conversational control

**Rescheduling philosophy**: rescheduling should be *fast* and *forgiving*—missing a block should not break the day; it should smoothly adapt.
- **Reschedule modes**:
  - **Automatic**: the system auto-shifts remaining blocks when you slip (within configured constraints).
  - **Approve-to-apply**: the system proposes the adjusted schedule; approval is a click or “confirm”.
- **Constraints** (configurable): don’t move immovable events, preserve sleep window, protect deep work blocks, respect “latest allowed time” per task.
- **Slip handling**: when a block runs over or is missed, prompt: *continue*, *skip*, *shorten*, *move later*, or *swap with next*.

### 2) Habits & routines
- define habits: target, schedule, measurement type (binary/count/minutes)
- streaks and adherence (but avoid streak obsession; focus on consistency)
- routines: morning/evening sequences
- habit-to-goal mapping (“this habit serves goal X”)
- enforcement options (opt-in): e.g. “don’t allow entertainment apps until workout complete”

### 3) Tasks, projects, and goals
- tasks: priority, due date, estimate, recurrence, dependencies
- projects: outcomes + next actions
- goals: 1–3 active goals + metrics + time horizon
- weekly planning: choose goals → select projects → select tasks → timebox
- review-driven backlog grooming (close loops and remove noise)

### 4) Time, scheduling, and focus
- calendar view (read-only first, then write)
- time blocks and timers (manual capture MVP; automation later)
- focus sessions (Pomodoro/deep work blocks)
- attention safeguards (later): app/site limits, “focus mode” profiles
- friction tools: “start ritual” prompts, reduce context switching

### 4.1) Passive activity capture (self-quantified layer)
**Goal**: reduce manual tracking by automatically capturing what you do, then turning it into usable time/accountability data.
- automatically capture active apps/windows and durations (ActivityWatch-like)
- classify activity into categories/projects (with human override)
- derive “focus” vs “context switching” metrics from window/app patterns
- connect passive activity → schedule blocks (did the block actually happen?)
- strict controls: pause/disable, private apps list, redaction, retention settings

### 5) Journal, mood/energy, and mental health support
- daily journal (free-form) + optional prompts (gratitude, wins, lessons)
- mood/energy check-ins (1–5) + tags (stress, anxiety, calm, focus)
- “CBT-lite” tools (opt-in): thought record, reframe, exposure ladder notes
- crisis/safety handling boundaries: the system should be supportive but not claim to be a therapist

### 6) Finances (personal CFO)
- transactions: manual first, then bank import (later)
- budgets and envelopes (simple)
- recurring bills and alerts
- goals: emergency fund, investment, debt payoff (optional scope)
- “behavioral finance” nudges: spending reflection, friction for impulse buys (opt-in)

**Investments (later)**:
- portfolio tracking (manual first; broker import later)
- risk profile + constraints (time horizon, max drawdown tolerance, ethical filters)
- automatic analysis: concentration risk, fees, diversification, drift vs target allocation
- suggestions: rebalancing, recurring purchases, tax-aware actions (where applicable)
- actionability ladder: insights → proposed orders → click-to-approve execution (integrations phase)

### 7) Reviews and coaching (the output engine)
**Daily (optional)**
- short recap + 1 suggestion

**Weekly (core)**
- adherence and progress toward goals
- time allocation vs priorities
- mood/energy trends
- what worked / what didn’t (from journal and outcomes)
- 1–3 recommended changes (small, high-leverage)
- draft plan for next week (you approve)

**Monthly/Quarterly**
- bigger strategy review, identity-level habits, environment redesign

### 7.1) Dashboards & gamification (progression system)
**Goal**: motivation through clarity and progression, without encouraging unhealthy streak obsession.
- **Stats/charts**: adherence, streaks (de-emphasized), time allocation, trend lines, “leading indicators” (sleep → focus, etc.)
- **Levels & XP (optional)**:
  - earn XP for completing planned blocks, habits, and key tasks (weighted by difficulty/impact)
  - **level** reflects consistency and challenge, not raw busyness
  - anti-gaming: cap repetitive low-value actions; reward “hard but important” work
- **Roadmap & objectives**:
  - clear objectives that **increase difficulty gradually**
  - “quest” style weekly themes (e.g. “Sleep stability week”, “Deep work week”) based on what you need now

### 8) Memory & personal model
Two layers:
- **Explicit memory** (you set): goals, constraints, preferences, schedules, values
- **Proposed memory** (assistant suggests): “You focus best 9–12”, “Meetings spike stress”

Tools:
- memory viewer/editor
- evidence-backed memory entries (show which data suggested it)
- delete/forget controls

### 9) Integrations & “hands” (later autonomy)
**Definition**: capabilities to act in external systems with your approval.
- calendar: create/reschedule events
- email/messages: draft, summarize, send (approval required)
- tasks: create tickets in external tools
- finance: categorize transactions, schedule bill reminders
- smart home / devices (optional)

Key constraints:
- least privilege scopes
- per-action approval (initially)
- policy-based approvals later (e.g. “auto-reschedule focus block within same day”)
- full audit trail

### 9.1) Autonomy maturity ladder (eventual direction)
**Goal**: “most things that can be automatic become automatic”, without losing safety or reversibility.
- **Level 0 — Manual**: you log; the system summarizes.
- **Level 1 — Passive sensing**: the system captures activity/biometrics/transactions; you label and correct.
- **Level 2 — Recommendations**: the system suggests schedule changes, habit tweaks, and finance improvements.
- **Level 3 — Propose-to-apply**: the system drafts changes and actions; you approve with one tap/confirm word.
- **Level 4 — Policy automation**: for safe scopes, actions auto-apply under policies (e.g. reschedule within constraints, rebalancing alerts).
- **Level 5 — Delegated autonomy**: broader automatic execution, still bounded by budgets, constraints, and an always-available kill switch.

---

## Evidence-based techniques to use (toolbox)
Jarvis should support multiple scientifically grounded approaches; you choose what fits.

### Behavior change & habit formation
- implementation intentions (“If X, then I will do Y”)
- tiny habits / minimum viable actions
- habit stacking (cue → routine → reward)
- environment design (reduce friction for good habits; add friction for bad)
- commitment devices (opt-in)

### Motivation and goal pursuit
- goal decomposition: outcome → milestones → process habits
- weekly planning ritual + review loop
- progress visualization that emphasizes trajectory (not perfection)

### Cognitive & emotional regulation (optional)
- journaling prompts
- cognitive reframing (CBT techniques, non-clinical framing)
- mindfulness / breathing micro-interventions
- sleep hygiene protocols

### Productivity & execution
- timeboxing and calendars that reflect priorities
- deep work blocks
- “next action” discipline
- batching and automation
- after-action reviews

### Social/identity reinforcement (optional)
- identity-based habit framing (“I am someone who…”)
- values clarification and trade-off checks
- “future self” letters / visualizations

---

## Safety, privacy, and trust model
This product touches your most sensitive data; trust is the feature.

### Data ownership
- export all data any time (JSON/CSV/Markdown)
- backup/restore
- clear retention policies

### LLM transparency
- always show what context is sent to the model (with redaction)
- cost tracking + rate limiting
- ability to run without LLM (basic deterministic insights)

### Default stance (current)
- prioritize assistant capability over maximum privacy, while keeping audit logs and controls
- plan a later “privacy-hardening” phase (more on-device processing, more summary-first prompts)

### Auditability
- “why did you suggest this?” explains:
  - inputs used
  - reasoning chain in plain language
  - confidence/uncertainty

### Autonomy safety rails
- default: no external actions
- later: approvals + policy system + scopes + logs
- emergency stop (“disable all automations”)

---

## Platform vision (Android + Linux)
### Android
- app-first, with an optional **launcher mode** later:
  - Today screen as home
  - quick capture from lock screen / notification shade
  - focus mode integration (app limits) if feasible

### Linux
- desktop app (or web app + wrapper) with:
  - fast keyboard-first capture
  - review writing/reading comfort
  - local backups and export workflows

### Cross-device sync (recommended)
- single-user identity; sync entries across devices
- conflict resolution for offline capture

---

## MVP → V1 → V2 roadmap (suggested)
This aligns with your existing MVP notes and expands forward.

### MVP (personal growth OS foundation)
- Android + Linux clients (capture + review reading)
- tracking: habits, tasks, schedule (timeline) + basic time blocks, journal, mood/energy
- dashboards: statistics/charts for gamification (streaks, adherence, time allocation, trends)
- weekly review generator (deterministic + optional LLM augmentation)
- minimal memory (pinned facts + suggested memory with approval)
- export + audit view for model context

### V1 (Life OS: stronger planning + stronger coaching)
- better planning flows: weekly plan builder + timeboxing + schedule optimization proposals
- coaching interventions library + personal experimentation (opt-in)
- integrations (read-only): calendar import, optional wearables

### V2 (“hands” and semi-autonomy)
- action integrations: calendar edits, email drafts/sends, task tool actions
- policy-based approvals
- proactive agent that runs “weekly planning” and proposes changes
- richer personal model and stable long-term memory

### V3 (finance + broader integrations)
- finance tracking + budgets (manual first; bank import later)
- integrations beyond calendar (email, messaging, task tools)

---

## UX requirements (high-level)
- **Fast capture**: common log flows < 10 seconds.
- **Low cognitive load**: minimal choices; smart defaults.
- **Review readability**: text-first with a few key numbers; charts optional.
- **Permission prompts**: clear, specific, reversible.
- **Accessible**: keyboard navigation, high contrast, screen reader support.

---

## Non-functional requirements
- **Reliability**: offline capture queue, safe sync, clear errors.
- **Security**: encrypt secrets/tokens at rest; least privilege scopes.
- **Privacy**: minimize data sent to models; support redaction.
- **Performance**: quick app startup; reviews generate within a reasonable budget.
- **Observability**: logs and an audit trail for computations and model calls.

---

## Open questions (to finalize the spec)
This section is intentionally explicit; answers will shape architecture and UX.

### Decisions captured (from your answers so far)
1) **Primary UX direction**: schedule/planning-first with a timeline; conversational mode later (hybrid).
2) **Hosting**: Go backend + Postgres, self-hosted on a VM; develop locally with a Postgres container.
3) **Privacy posture**: capability-first; add privacy-hardening later.
4) **Enforcement**: adaptive coaching + adaptive friction; more push when needed, less when not.
5) **Roadmap priority**: habits + tasks + schedule/planning + reviews first; finance/integrations later.

### Remaining open questions (high leverage)
1) What does “schedule” mean for you: strict calendar blocks, flexible task list, or a mix?
2) Rescheduling should support **both**: auto-reschedule and approve-to-apply, with easy manual adjustments.
3) Gamification should include **stats/charts** and optionally **levels/progression**, tuned to techniques that improve long-term consistency.


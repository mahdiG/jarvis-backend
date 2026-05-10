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
    ID          uint           `gorm:"primaryKey"`
    Title       string
    Description string
    ParentID    *uint
    Status      string         // active, done, archived
    Type        string         // "value", "goal", "roadmap", "task", "habit"
    ScheduledFor *time.Time
    Metadata    datatypes.JSON `gorm:"type:json"` // e.g., {"ai_breakdown": [...], "score": 0.7}
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

AI SAYS: And a content note
This exact decision—"why I used a single recursive table instead of premature micro-categorization for my AI Life OS"—is the kind of piece that resonates deeply with senior engineers. It shows you know how to balance immediate delivery with long-term flexibility, and that you're not afraid to start simple. Write it down now, even as a paragraph in your progress log.

Fighting temptations to use new shiny technologies and stay with boring mature tech that works and ships fast.
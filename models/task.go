package models

type Task struct {
	Base
	Title       string `gorm:"not null" validate:"required"`
	Description string
	ParentID    UID
	// Status       string // active, done, archived
	// Type         string // "value", "goal", "roadmap", "task", "habit"
	// ScheduledFor *time.Time
	// Metadata     datatypes.JSON `gorm:"type:json"` // e.g., {"ai_breakdown": [...], "score": 0.7}
}

package models

import "time"

type Task struct {
	Base
	Title       string `gorm:"not null" validate:"required"`
	Description string
	ParentID    UID
	// Target score for binary (done/undone) task is 1
	TargetScore   float64 `gorm:"default:1" validate:"min=0"`
	Score         float64 `gorm:"default:0"`
	ScoreUnit     string
	ScheduledFrom *time.Time
	ScheduledTo   *time.Time
	DoneAt        *time.Time
	Tags          []Tag `gorm:"many2many:task_tags;"`
}

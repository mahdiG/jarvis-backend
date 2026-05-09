package models

// Habit represents a tracked habit in the system.
type Habit struct {
	Base
	Name string `gorm:"not null"`
}

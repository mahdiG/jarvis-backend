package models

type Conversation struct {
	Base
	Title string `gorm:"not null" validate:"required"`
}
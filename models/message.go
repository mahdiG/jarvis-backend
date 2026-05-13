package models

type Message struct {
	Base
	ConversationID UID    `gorm:"not null;index" validate:"required"`
	Role           string `gorm:"not null" validate:"required"`
	Content        string `gorm:"not null" validate:"required"`
}

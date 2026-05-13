package repositories

import (
	"jarvis/models"

	"gorm.io/gorm/clause"
)

func GetMessages(conversationID models.UID) ([]models.Message, error) {
	var messages []models.Message

	result := db.
		Where("conversation_id = ?", conversationID).
		Order("created_at asc").
		Find(&messages)

	return messages, result.Error
}

func GetMessage(condition models.Message) (models.Message, error) {
	var message models.Message

	result := db.
		Clauses(clause.Returning{}).
		Where(&condition).
		First(&message)

	return message, result.Error
}

func CreateMessage(message models.Message) (models.Message, error) {
	result := db.Create(&message)
	return message, result.Error
}
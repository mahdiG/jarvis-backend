package repositories

import (
	"jarvis/models"

	"gorm.io/gorm/clause"
)

func GetConversations(limit, offset int) ([]models.Conversation, error) {
	var conversations []models.Conversation

	query := db.Model(&models.Conversation{})
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	result := query.
		Preload("Messages").
		Find(&conversations)

	return conversations, result.Error
}

func GetConversation(condition models.Conversation) (models.Conversation, error) {
	var conversation models.Conversation

	result := db.
		Clauses(clause.Returning{}).
		Preload("Messages").
		Where(&condition).
		First(&conversation)

	return conversation, result.Error
}

func CreateConversation(conversation models.Conversation) (models.Conversation, error) {
	result := db.Create(&conversation)
	return conversation, result.Error
}

func UpdateConversation(conversation models.Conversation) (models.Conversation, error) {
	result := db.
		Clauses(clause.Returning{}).
		Where("id = ?", conversation.ID).
		Updates(&conversation)

	if result.Error != nil {
		return conversation, result.Error
	}
	if result.RowsAffected == 0 {
		return conversation, ErrRecordNotFound
	}
	return conversation, nil
}

func DeleteConversation(condition models.Conversation) error {
	result := db.
		Where(&condition).
		Delete(&models.Conversation{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
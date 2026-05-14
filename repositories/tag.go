package repositories

import (
	"jarvis/models"

	"gorm.io/gorm/clause"
)

func GetTags(limit, offset int) ([]models.Tag, error) {
	var tags []models.Tag

	query := db.Model(&models.Tag{})
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	result := query.Find(&tags)

	return tags, result.Error
}

func GetTag(condition models.Tag) (models.Tag, error) {
	var tag models.Tag

	result := db.
		Clauses(clause.Returning{}).
		Where(&condition).
		First(&tag)

	return tag, result.Error
}

func CreateTag(tag models.Tag) (models.Tag, error) {
	result := db.Create(&tag)
	return tag, result.Error
}

func UpdateTag(tag models.Tag) (models.Tag, error) {
	result := db.
		Clauses(clause.Returning{}).
		Where("id = ?", tag.ID).
		Updates(&tag)

	if result.Error != nil {
		return tag, result.Error
	}
	if result.RowsAffected == 0 {
		return tag, ErrRecordNotFound
	}
	return tag, nil
}

func DeleteTag(condition models.Tag) error {
	result := db.
		Where(&condition).
		Delete(&models.Tag{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
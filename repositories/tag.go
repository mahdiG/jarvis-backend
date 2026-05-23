package repositories

import (
	"jarvis/models"

	"gorm.io/gorm"
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

func CreateTags(tags []models.Tag) ([]models.Tag, error) {
	result := db.Create(&tags)
	return tags, result.Error
}

// UpdateTags performs a batch partial update. Each tag in the slice must have
// a non‑empty ID. If any tag is not found the entire update is rolled back and
// ErrRecordNotFound is returned.
func UpdateTags(tags []models.Tag) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			result := tx.
				Model(&models.Tag{}).
				Where("id = ?", tag.ID).
				Updates(&tag)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrRecordNotFound
			}
		}
		return nil
	})
}

// DeleteTags deletes tags by their IDs. Non‑existent IDs are silently ignored.
func DeleteTags(ids []models.UID) error {
	return db.Where("id IN ?", ids).Delete(&models.Tag{}).Error
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

package repositories

import (
	"jarvis/models"

	"gorm.io/gorm/clause"
)

func GetNotes(limit, offset int) ([]models.Note, error) {
	var notes []models.Note

	query := db.Model(&models.Note{})
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	result := query.
		Preload("Tags").
		Find(&notes)

	return notes, result.Error
}

func GetNote(condition models.Note) (models.Note, error) {
	var note models.Note

	result := db.
		Clauses(clause.Returning{}).
		Preload("Tags").
		Where(&condition).
		First(&note)

	return note, result.Error
}

func CreateNote(note models.Note) (models.Note, error) {
	result := db.Create(&note)
	return note, result.Error
}

func UpdateNote(note models.Note) (models.Note, error) {
	result := db.
		Clauses(clause.Returning{}).
		Where("id = ?", note.ID).
		Updates(&note)

	if result.Error != nil {
		return note, result.Error
	}
	if result.RowsAffected == 0 {
		return note, ErrRecordNotFound
	}
	return note, nil
}

func GetTrashNotes(limit, offset int) ([]models.Note, error) {
	var notes []models.Note

	query := db.Unscoped().Model(&models.Note{}).Where("deleted_at IS NOT NULL")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	result := query.
		Preload("Tags").
		Find(&notes)

	return notes, result.Error
}

func RestoreNote(id models.UID) (models.Note, error) {
	var note models.Note

	result := db.Unscoped().
		Model(&models.Note{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)

	if result.Error != nil {
		return note, result.Error
	}
	if result.RowsAffected == 0 {
		return note, ErrRecordNotFound
	}

	// Fetch the restored note with its preloaded associations.
	getResult := db.
		Preload("Tags").
		First(&note, "id = ?", id)
	if getResult.Error != nil {
		return note, getResult.Error
	}

	return note, nil
}

func PermanentDeleteNote(id models.UID) error {
	result := db.Unscoped().
		Where("id = ?", id).
		Delete(&models.Note{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func DeleteNote(condition models.Note) error {
	result := db.
		Where(&condition).
		Delete(&models.Note{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

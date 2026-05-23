package repositories

import (
	"jarvis/models"
	"log/slog"

	"gorm.io/gorm"
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

func CreateNotes(notes []models.Note) ([]models.Note, error) {
	result := db.Create(&notes)
	return notes, result.Error
}

// UpdateNotes performs a batch partial update. Each note in the slice must have
// a non‑empty ID. If any note is not found the entire update is rolled back and
// ErrRecordNotFound is returned.
func UpdateNotes(notes []models.Note) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, note := range notes {
			result := tx.
				Model(&models.Note{}).
				Where("id = ?", note.ID).
				Updates(&note)
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

// SoftDeleteNotes soft‑deletes the notes with the given IDs. Non‑existent IDs
// are silently ignored.
func SoftDeleteNotes(ids []models.UID) error {
	return db.Where("id IN ?", ids).Delete(&models.Note{}).Error
}

// RestoreNotes restores soft‑deleted notes by ID and returns the restored notes.
// Returns ErrRecordNotFound if none of the IDs matched a soft‑deleted note.
func RestoreNotes(ids []models.UID) ([]models.Note, error) {
	result := db.Unscoped().
		Model(&models.Note{}).
		Where("id IN ? AND deleted_at IS NOT NULL", ids).
		Update("deleted_at", nil)

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrRecordNotFound
	}

	var notes []models.Note
	err := db.
		Preload("Tags").
		Where("id IN ?", ids).
		Find(&notes).Error
	return notes, err
}

// HardDeleteNotes permanently deletes notes by ID (regardless of soft‑delete
// status). Non‑existent IDs are silently ignored.
func HardDeleteNotes(ids []models.UID) error {
	slog.Debug("Hard delete notes", "ids", ids)

	var notes []models.Note
	result := db.Unscoped().
		Where("id IN ?", ids).
		Find(&notes)

	if result.Error != nil {
		return result.Error
	}
	if len(notes) == 0 {
		slog.Debug("no notes to delete")
		return nil
	}

	result = db.Unscoped().
		Select("Tags", "Files").
		Delete(&notes)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

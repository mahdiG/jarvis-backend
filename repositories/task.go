package repositories

import (
	"jarvis/models"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetTasks(limit, offset int) ([]models.Task, error) {
	var tasks []models.Task

	query := db.
		Model(&models.Task{}).
		Preload("Tags").
		Preload("Children.Children.Children").
		Where("parent_id", nil)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	result := query.Find(&tasks)

	return tasks, result.Error
}

func GetTask(condition models.Task) (models.Task, error) {
	var task models.Task

	result := db.
		Clauses(clause.Returning{}).
		Preload("Tags").
		Preload("Children.Children.Children").
		Where(&condition).
		First(&task)

	return task, result.Error
}

func CreateTasks(tasks []models.Task) ([]models.Task, error) {
	result := db.Create(&tasks)
	return tasks, result.Error
}

// UpdateTasks performs a batch partial update on tasks including their
// associations (Tags, Children).  Uses GORM's Save() with FullSaveAssociations
// so that zero values (0, nil, "") are correctly written to the database —
// e.g. setting Score = 0 or DoneAt = nil to mark a task as undone.  If the
// task does not exist it is created (upsert behaviour).
func UpdateTasks(tasks []models.Task) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, task := range tasks {
			result := tx.
				Session(&gorm.Session{FullSaveAssociations: true}).
				Save(&task)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
}

func GetTrashTasks(limit, offset int) ([]models.Task, error) {
	var tasks []models.Task

	query := db.Model(&models.Task{})
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	result := query.
		Unscoped().
		Where("deleted_at IS NOT NULL").
		Preload("Tags").
		Preload("Children.Children.Children").
		Find(&tasks)

	return tasks, result.Error
}

// SoftDeleteTasks soft‑deletes the tasks with the given IDs. Non‑existent IDs
// are silently ignored.
func SoftDeleteTasks(ids []models.UID) error {
	return db.Where("id IN ?", ids).Delete(&models.Task{}).Error
}

// RestoreTasks restores soft‑deleted tasks by ID and returns the restored tasks.
// Returns ErrRecordNotFound if none of the IDs matched a soft‑deleted task.
func RestoreTasks(ids []models.UID) ([]models.Task, error) {
	result := db.Unscoped().
		Model(&models.Task{}).
		Where("id IN ? AND deleted_at IS NOT NULL", ids).
		Update("deleted_at", nil)

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrRecordNotFound
	}

	var tasks []models.Task
	err := db.
		Preload("Children.Children.Children").
		Where("id IN ?", ids).
		Find(&tasks).Error
	return tasks, err
}

// HardDeleteTasks permanently deletes tasks by ID (regardless of soft‑delete
// status). Non‑existent IDs are silently ignored.
func HardDeleteTasks(ids []models.UID) error {
	slog.Debug("Hard delete tasks", "ids", ids)

	var tasks []models.Task
	result := db.Unscoped().
		Where("id IN ?", ids).
		Find(&tasks)

	if result.Error != nil {
		return result.Error
	}
	if len(tasks) == 0 {
		slog.Debug("no tasks to delete")
		return nil
	}

	result = db.Unscoped().
		Select("Tags").
		Delete(&tasks)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
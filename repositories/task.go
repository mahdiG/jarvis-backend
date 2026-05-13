package repositories

import (
	"jarvis/models"

	"gorm.io/gorm/clause"
)

func GetTasks(limit, offset int) ([]models.Task, error) {
	var tasks []models.Task

	query := db.Model(&models.Task{})
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
		Where(&condition).
		First(&task)

	return task, result.Error
}

func CreateTask(task models.Task) (models.Task, error) {
	result := db.Create(&task)
	return task, result.Error
}

func UpdateTask(task models.Task) (models.Task, error) {
	result := db.
		Clauses(clause.Returning{}).
		Where("id = ?", task.ID).
		Updates(&task)

	if result.Error != nil {
		return task, result.Error
	}
	if result.RowsAffected == 0 {
		return task, ErrRecordNotFound
	}
	return task, nil
}

func DeleteTask(condition models.Task) error {
	result := db.
		Where(&condition).
		Delete(&models.Task{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

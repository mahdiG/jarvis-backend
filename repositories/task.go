package repositories

import (
	"jarvis/models"

	"gorm.io/gorm/clause"
)

func GetTasks() ([]models.Task, error) {
	var tasks []models.Task

	result := db.
		Find(&tasks)

	return tasks, result.Error
}

func GetTask(id models.UID) (models.Task, error) {
	var task models.Task

	result := db.
		Clauses(clause.Returning{}).
		Where(models.Task{Base: models.Base{ID: id}}).
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

func DeleteTask(id models.UID) error {
	result := db.
		Where(models.Task{Base: models.Base{ID: id}}).
		Delete(&models.Task{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

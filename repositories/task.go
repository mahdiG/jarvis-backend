package repositories

import (
	"jarvis/models"
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
		Where(models.Task{Base: models.Base{ID: id}}).
		First(&task, id)

	// return task, utils.WrapError(result.Error)
	return task, result.Error
}

func CreateTask(task *models.Task) error {
	uid, err := models.NewUID()
	if err != nil {
		return err
	}

	task.ID = uid

	result := db.Create(task)
	return result.Error
}

func UpdateTask(id models.UID, updates map[string]any) (models.Task, error) {
	var task models.Task

	result := db.
		Model(&models.Task{Base: models.Base{ID: id}}).
		Updates(updates).
		First(&task)

	return task, result.Error
}

func DeleteTask(id models.UID) error {
	result := db.
		Where(models.Task{Base: models.Base{ID: id}}).
		Delete(&models.Task{})

	return result.Error
}

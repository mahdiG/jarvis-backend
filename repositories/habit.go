package repositories

import (
	"jarvis/models"
)

func GetHabit(id models.UID) (models.Habit, error) {
	var habit models.Habit

	result := db.
		Where(models.Habit{Base: models.Base{ID: id}}).
		First(&habit, id)

	// return habit, utils.WrapError(result.Error)
	return habit, result.Error
}

package repositories

import (
	"jarvis/models"
)

func GetHabit(id models.UID) (models.Habit, error) {
	var habit models.Habit

	result := db.
		Where(models.Habit{Base: models.Base{ID: id}}).
		First(&habit, id)

	if result.Error != nil {
		// Should we log error here?
		// Should we return error nil and an empty habit if not found without logging error?
		return habit, result.Error
	}

	return habit, nil
}

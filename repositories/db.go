package repositories

import (
	"gorm.io/gorm"
)

var db *gorm.DB

// Init sets the global database connection used by all repository functions.
func Init(database *gorm.DB) {
	db = database
}

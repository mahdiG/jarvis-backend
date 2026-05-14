package models

type Tag struct {
	Base
	Name string `gorm:"uniqueIndex;not null" validate:"required"`
}

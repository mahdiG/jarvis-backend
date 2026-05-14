package models

type Tag struct {
	Base
	Name string `gorm:"not null" validate:"required"`
}

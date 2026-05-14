package models

type File struct {
	Base
	Name string `gorm:"not null"`
	Url  string `gorm:"not null"`
}

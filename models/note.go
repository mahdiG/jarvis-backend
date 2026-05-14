package models

type Note struct {
	Base
	Title   string `gorm:"not null" validate:"required"`
	Content string
	Tags    []Tag  `gorm:"many2many:note_tags;"`
	Files   []File `gorm:"many2many:note_files;"`
}
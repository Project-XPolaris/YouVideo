package database

import "gorm.io/gorm"

type Video struct {
	gorm.Model
	Path      string
	LibraryId uint
	Cover     string
}

package database

import "gorm.io/gorm"

type Folder struct {
	gorm.Model
	Path      string
	Videos    []*Video
	LibraryId uint
}

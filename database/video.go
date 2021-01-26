package database

import "gorm.io/gorm"

type Video struct {
	gorm.Model
	Files     []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name      string
	LibraryId uint
	BaseDir   string
}

package database

import "gorm.io/gorm"

type Video struct {
	gorm.Model
	Files     []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name      string
	LibraryId uint
	BaseDir   string
	Tags      []*Tag `gorm:"many2many:video_tags;"`
	History   []*History
}

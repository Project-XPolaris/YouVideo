package database

import (
	"gorm.io/gorm"
	"time"
)

type Video struct {
	gorm.Model
	Files     []File `gorm:"constraint:OnDelete:CASCADE;"`
	Name      string
	LibraryId uint
	Library   *Library
	BaseDir   string
	Tags      []*Tag           `gorm:"many2many:video_tags;"`
	Type      string           `gorm:"default:video"`
	History   []*History       `gorm:"constraint:OnDelete:CASCADE;"`
	Infos     []*VideoMetaItem `gorm:"many2many:video_infos;"`
	FolderID  *uint
	SubjectId *uint
	Release   *time.Time
	EntityID  *uint
}

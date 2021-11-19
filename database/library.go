package database

import "gorm.io/gorm"

type Library struct {
	gorm.Model
	Path             string
	Name             string
	DefaultVideoType string  `gorm:"default:video"`
	Videos           []Video `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Users            []*User `gorm:"many2many:library_users;"`
	Folders          []*Folder
}

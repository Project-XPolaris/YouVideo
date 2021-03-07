package database

import "gorm.io/gorm"

type Library struct {
	gorm.Model
	Path   string
	Name   string
	Videos []Video `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

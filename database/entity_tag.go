package database

import "gorm.io/gorm"

type EntityTag struct {
	gorm.Model
	Name     string
	Value    string
	Entities []*Entity `gorm:"many2many:entity_entity_tags;"`
}

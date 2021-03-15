package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Uid     string
	Library []*Library `gorm:"many2many:library_users;"`
}

package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Uid     string     `gorm:"unique"`
	Library []*Library `gorm:"many2many:library_users;"`
	Tags    []*Tag     `gorm:"many2many:tag_users;"`
}

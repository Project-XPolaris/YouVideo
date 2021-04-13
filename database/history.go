package database

import "gorm.io/gorm"

type History struct {
	gorm.Model
	UserID  uint
	VideoID uint
	User    *User
	Video   *Video
}

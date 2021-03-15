package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Instance *gorm.DB
)

func Connect() error {
	var err error
	Instance, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	Instance.AutoMigrate(&Video{}, &Library{}, &File{}, &Tag{}, &User{})
	return nil
}

func InitDatabase() error {
	var user User
	return Instance.FirstOrCreate(&user, User{Uid: "-1"}).Error
}

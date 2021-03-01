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
	Instance.AutoMigrate(&Video{}, &Library{}, &File{}, &Tag{})
	return nil
}

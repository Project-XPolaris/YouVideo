package database

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Instance *gorm.DB
	Logger   = log.New().WithFields(log.Fields{
		"scope": "Database",
	})
)

func Connect() error {
	var err error
	Instance, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		Logger.Errorln(err)
		return err
	}
	Instance.AutoMigrate(&Video{}, &Library{})
	Logger.Info("database connected")
	return nil
}

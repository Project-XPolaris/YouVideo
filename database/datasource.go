package database

import (
	"github.com/allentom/harukap/datasource"
	"gorm.io/gorm"
)

var DefaultPlugin = &datasource.Plugin{
	OnConnected: func(db *gorm.DB) {
		Instance = db
		Instance.AutoMigrate(&Video{}, &Library{}, &File{}, &Tag{}, &User{}, &History{}, &Folder{})
		var user User
		Instance.FirstOrCreate(&user, User{Uid: "-1"})
	},
}

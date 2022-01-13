package database

import (
	"gorm.io/gorm"
)

type VideoMetaItem struct {
	gorm.Model
	Key    string   `json:"key"`
	Value  string   `json:"value"`
	Videos []*Video `gorm:"many2many:video_infos;"`
}

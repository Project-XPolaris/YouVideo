package database

import "gorm.io/gorm"

type File struct {
	gorm.Model
	Path           string
	VideoId        uint
	Cover          string
	Duration       float64
	Size           int64
	Bitrate        int64
	MainVideoCodec string
	MainAudioCodec string
}

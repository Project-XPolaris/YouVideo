package database

import "gorm.io/gorm"

type Video struct {
	gorm.Model
	Path           string
	LibraryId      uint
	Cover          string
	Duration       float64
	Size           int64
	Bitrate        int64
	MainVideoCodec string
	MainAudioCodec string
}

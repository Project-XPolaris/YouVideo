package database

import (
	"gorm.io/gorm"
	"strings"
)

type Subtitles struct {
	gorm.Model
	FileId uint
	Path   string
	Label  string
}

func ReadOrCreateSubtitles(path string, fileId uint) (*Subtitles, error) {
	parts := strings.Split(path, ".")
	subtitles := &Subtitles{
		Path:   path,
		FileId: fileId,
	}
	if len(parts) > 2 {
		subtitles.Label = parts[len(parts)-2]
	}
	err := Instance.FirstOrCreate(subtitles, Subtitles{Path: path}).Error
	return subtitles, err
}

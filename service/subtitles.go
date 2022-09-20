package service

import (
	"github.com/asticode/go-astisub"
	"github.com/projectxpolaris/youvideo/database"
	"time"
)

type CC struct {
	Index     int
	StartTime time.Duration
	EndTime   time.Duration
	Text      string
}

func GetCloseCaption(file *database.File) ([]*CC, error) {
	var cc []*CC
	source, err := astisub.OpenFile(file.Subtitles[0].Path)
	if err != nil {
		return cc, err
	}
	for _, item := range source.Items {
		cc = append(cc, &CC{
			StartTime: item.StartAt,
			EndTime:   item.EndAt,
			Text:      item.Lines[0].Items[0].Text,
			Index:     item.Index,
		})

	}
	return cc, nil
}

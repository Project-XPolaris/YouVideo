package service

import (
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

var DefaultVideoCoverGenerator *VideoCoverMetaAnalyzer

func init() {
	DefaultVideoCoverGenerator = &VideoCoverMetaAnalyzer{
		In: make(chan *database.File, 100000),
		Logger: logrus.WithFields(logrus.Fields{
			"scope": "DefaultVideoCoverGenerator",
		}),
	}
	go DefaultVideoCoverGenerator.Run()
}

type VideoCoverMetaAnalyzer struct {
	In     chan *database.File
	Logger *logrus.Entry
}

func (a *VideoCoverMetaAnalyzer) Run() {
	for true {
		file := <-a.In
		fileLogger := a.Logger.WithFields(logrus.Fields{
			"path": file.Path,
			"file": filepath.Base(file.Path),
		})
		coverPath, err := GenerateVideoCover(file.Path)
		if err != nil {
			fileLogger.Error(err)
		} else {
			file.AutoGenCover = true
			file.Cover = filepath.Join(config.Instance.CoversStore, filepath.Base(coverPath))
		}
		err = database.Instance.Save(file).Error
		if err != nil {
			fileLogger.Error(err)
		}
	}
}

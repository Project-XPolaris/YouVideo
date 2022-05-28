package service

import (
	"github.com/projectxpolaris/youvideo/database"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
)

var DefaultVideoMetaAnalyzer *VideoMetaAnalyzer

func init() {
	DefaultVideoMetaAnalyzer = &VideoMetaAnalyzer{
		In: make(chan VideoMetaAnalyzerInput, 100000),
		Logger: logrus.WithFields(logrus.Fields{
			"scope": "DefaultVideoMetaAnalyzer",
		}),
	}
	go DefaultVideoMetaAnalyzer.Run()
}

type VideoMetaAnalyzerInput struct {
	File    *database.File
	OnDone  chan struct{}
	OnError chan error
}
type VideoMetaAnalyzer struct {
	In     chan VideoMetaAnalyzerInput
	Logger *logrus.Entry
}

func (a *VideoMetaAnalyzer) Run() {

	for true {
		input := <-a.In
		file := input.File
		fileLogger := a.Logger.WithFields(logrus.Fields{
			"path": file.Path,
			"file": filepath.Base(file.Path),
		})
		// get meta data
		meta, err := GetVideoFileMeta(file.Path)
		if err != nil {
			if input.OnError != nil {
				input.OnError <- err
			}
			fileLogger.Error(err)
			return
		}
		file.Duration = meta.Format.DurationSeconds
		size, err := strconv.ParseInt(meta.Format.Size, 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Size = size
		}

		bitrate, err := strconv.ParseInt(meta.Format.BitRate, 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Bitrate = bitrate
		}

		// parse stream
		for _, stream := range meta.Streams {
			if stream.CodecType == "video" && len(file.MainVideoCodec) == 0 {
				file.MainVideoCodec = stream.CodecName
				continue
			}
			if stream.CodecType == "audio" && len(file.MainAudioCodec) == 0 {
				file.MainAudioCodec = stream.CodecName
			}
		}

		err = database.Instance.Save(file).Error
		if err != nil {
			if input.OnError != nil {
				input.OnError <- err
			}
			fileLogger.Error(err)
		}
		if input.OnDone != nil {
			close(input.OnDone)
		}
	}
}

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
		duration, err := strconv.ParseFloat(meta.GetFormat().GetDuration(), 10)
		if err != nil {
			fileLogger.Error(err)
		} else {
			file.Duration = duration
		}
		size, err := strconv.ParseInt(meta.GetFormat().GetSize(), 10, 64)
		if err != nil {
			fileLogger.Error(err)
		} else {
			file.Size = size
		}

		bitrate, err := strconv.ParseInt(meta.GetFormat().GetBitRate(), 10, 64)
		if err != nil {
			fileLogger.Error(err)
		} else {
			file.Bitrate = bitrate
		}

		// parse stream
		for _, stream := range meta.GetStreams() {
			if stream.GetCodecType() == "video" && len(file.MainVideoCodec) == 0 {
				file.MainVideoCodec = stream.GetCodecName()
				continue
			}
			if stream.GetCodecType() == "audio" && len(file.MainAudioCodec) == 0 {
				file.MainAudioCodec = stream.GetCodecName()
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

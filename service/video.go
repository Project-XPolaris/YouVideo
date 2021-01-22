package service

import (
	"github.com/projectxpolaris/youvideo/database"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

var VideoLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "Service.Video",
})

func ScanVideo(library *database.Library) error {
	targetExtensions := []string{
		"mp4", "mkv",
	}
	err := afero.Walk(AppFs, library.Path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, extension := range targetExtensions {
			if strings.HasSuffix(info.Name(), extension) {
				err := CreateVideo(path, library.ID)
				if err != nil {
					logrus.Error(err)
				}
				break
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return err
}

func CreateVideo(path string, libraryId uint) error {
	var recordCount int64
	err := database.Instance.Model(&database.Video{}).Where("path = ?", path).Count(&recordCount).Error
	if err != nil {
		return err
	}
	if recordCount != 0 {
		VideoLogger.WithFields(logrus.Fields{
			"path": path,
		}).Warn("video index exist!")
		return nil
	}
	video := &database.Video{
		Path:      path,
		LibraryId: libraryId,
	}
	coverPath, err := GenerateVideoCover(path)
	if err != nil {
		logrus.Error(err)
	} else {
		video.Cover = filepath.Base(coverPath)
	}
	err = database.Instance.Save(video).Error
	return err
}

type VideoQueryOption struct {
	Page     int
	PageSize int
}

func GetVideoList(option VideoQueryOption) (int64, []database.Video, error) {
	var result []database.Video
	var count int64
	queryBuilder := database.Instance.Model(&database.Video{})
	err := queryBuilder.Limit(option.PageSize).Count(&count).Offset((option.Page - 1) * option.PageSize).Find(&result).Error
	return count, result, err
}

func GetVideoById(id uint) (*database.Video, error) {
	var video database.Video
	err := database.Instance.Find(&video, id).Error
	return &video, err
}

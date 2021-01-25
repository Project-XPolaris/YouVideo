package service

import (
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strconv"
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

	// get meta data
	meta, metaerr := GetVideoFileMeta(path)
	if metaerr == nil {
		duration, err := strconv.ParseFloat(meta.GetFormat().GetDuration(), 10)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			video.Duration = duration
		}

		size, err := strconv.ParseInt(meta.GetFormat().GetSize(), 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			video.Size = size
		}

		bitrate, err := strconv.ParseInt(meta.GetFormat().GetBitRate(), 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			video.Bitrate = bitrate
		}

		// parse stream
		for _, stream := range meta.GetStreams() {
			if stream.GetCodecType() == "video" && stream.GetProfile() == "Main" {
				video.MainVideoCodec = stream.GetCodecName()
				continue
			}
			if stream.GetCodecType() == "audio" && len(video.MainAudioCodec) == 0 {
				video.MainAudioCodec = stream.GetCodecName()
			}

		}
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

func DeleteVideoById(id uint) error {
	var video database.Video
	err := database.Instance.First(&video, id).Error
	if err != nil {
		return err
	}
	err = database.Instance.
		Model(&database.Video{}).
		Unscoped().
		Where("id = ?", id).
		Delete(&database.Video{}).
		Error
	if err != nil {
		return err
	}
	if len(video.Cover) > 0 {
		coverPath := filepath.Join(config.AppConfig.CoversStore, video.Cover)
		if _, err = os.Stat(coverPath); err == nil {
			err = os.Remove(coverPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

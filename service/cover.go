package service

import (
	"context"
	"fmt"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
	"github.com/sirupsen/logrus"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"
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

func getImageDimension(imagePath string) (int, int, error) {
	buff, err := plugin.ReadFileBuffer(imagePath)
	if err != nil {
		return 0, 0, err
	}
	targetImage, _, err := image.DecodeConfig(buff)
	if err != nil {
		return 0, 0, err
	}
	return targetImage.Width, targetImage.Height, nil
}
func GetTargetCover(targetFilePath string) string {
	baseDir := filepath.Dir(targetFilePath)
	fileExt := filepath.Ext(targetFilePath)
	videoName := strings.TrimSuffix(filepath.Base(targetFilePath), fileExt)
	targetCoverFilePaths := []string{
		"cover.jpg",
		"cover.png",
		"cover.jpeg",
		"cover.JPEG",
		"cover.PNG",
		fmt.Sprintf("%s.jpg", videoName),
		fmt.Sprintf("%s.png", videoName),
		fmt.Sprintf("%s.jpeg", videoName),
		fmt.Sprintf("%s.JPEG", videoName),
		fmt.Sprintf("%s.PNG", videoName),
	}
	for _, targetCoverFilePath := range targetCoverFilePaths {
		coverSourcePath := filepath.Join(baseDir, targetCoverFilePath)
		if util.CheckFileExist(coverSourcePath) {
			return coverSourcePath
		}
	}
	return ""
}
func (a *VideoCoverMetaAnalyzer) GenerateFromVideoShot() {

}
func (a *VideoCoverMetaAnalyzer) Run() {
	for true {
		file := <-a.In
		fileLogger := a.Logger.WithFields(logrus.Fields{
			"path": file.Path,
			"file": filepath.Base(file.Path),
		})
		coverFilePath := GetTargetCover(file.Path)
		// remove cover is not exist
		if len(file.Cover) > 0 {
			existCoverPath := filepath.Join(config.Instance.CoversStore, file.Cover)
			storage := plugin.GetDefaultStorage()
			isExist, _ := storage.IsExist(context.Background(), plugin.GetDefaultBucket(), existCoverPath)
			if !isExist {
				file.Cover = ""
			}
		}
		if len(file.Cover) > 0 && len(coverFilePath) == 0 {
			continue
		}
		// check if cover is same as old
		if len(coverFilePath) > 0 {
			// check if cover is same as old
			md5, err := util.MD5Checksum(coverFilePath)
			if err != nil {
				fileLogger.Error(err)
				continue
			}
			// if cover is same as old, skip
			if file.CoverFileMd5 == md5 {
				continue
			}
			file.CoverFileMd5 = md5
		}
		// use video shot
		if len(coverFilePath) == 0 {
			coverPath, err := GenerateVideoCover(file.Path)
			if err != nil {
				fileLogger.Error(err)
				continue
			}
			file.AutoGenCover = true
			file.Cover = filepath.Base(coverPath)
		} else {
			thumbnailFileName, err := GenerateThumbnail(coverFilePath)
			if err != nil {
				fileLogger.Error(err)
				continue
			}
			file.Cover = thumbnailFileName
			file.AutoGenCover = false
		}
		coverStorePath := filepath.Join(config.Instance.CoversStore, file.Cover)
		width, height, err := getImageDimension(coverStorePath)
		if err != nil {
			fileLogger.Error(err)
			continue
		}
		file.CoverWidth = int64(width)
		file.CoverHeight = int64(height)
		err = database.Instance.Save(file).Error
		if err != nil {
			fileLogger.Error(err)
		}
	}
}

func GenerateImageCover(file *database.File) error {
	coverFilePath := GetTargetCover(file.Path)
	// remove cover is not exist
	if len(file.Cover) > 0 {
		existCoverPath := filepath.Join(config.Instance.CoversStore, file.Cover)
		storage := plugin.GetDefaultStorage()
		isExist, _ := storage.IsExist(context.Background(), plugin.GetDefaultBucket(), existCoverPath)
		if !isExist {
			file.Cover = ""
		}
	}
	if len(file.Cover) > 0 && len(coverFilePath) == 0 {
		return nil
	}
	// check if cover is same as old
	if len(coverFilePath) > 0 {
		// check if cover is same as old
		md5, err := util.MD5Checksum(coverFilePath)
		if err != nil {
			return err
		}
		// if cover is same as old, skip
		if file.CoverFileMd5 == md5 {
			return nil
		}
		file.CoverFileMd5 = md5
	}
	// use video shot
	if len(coverFilePath) == 0 {
		coverPath, err := GenerateVideoCover(file.Path)
		if err != nil {
			return err
		}
		file.AutoGenCover = true
		file.Cover = filepath.Base(coverPath)
	} else {
		thumbnailFileName, err := GenerateThumbnail(coverFilePath)
		if err != nil {
			return err
		}
		file.Cover = thumbnailFileName
		file.AutoGenCover = false
	}
	coverStorePath := filepath.Join(config.Instance.CoversStore, file.Cover)
	width, height, err := getImageDimension(coverStorePath)
	if err != nil {
		return err
	}
	file.CoverWidth = int64(width)
	file.CoverHeight = int64(height)
	return nil
}

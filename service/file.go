package service

import (
	"errors"
	"fmt"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/youtrans"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func RemoveFileById(id uint) error {
	var file database.File
	err := database.Instance.Model(&database.File{}).Where("id = ?", id).First(&file).Error
	if err != nil {
		return err
	}
	if len(file.Cover) > 0 {
		coverPath := filepath.Join(config.AppConfig.CoversStore, file.Cover)
		if _, err = os.Stat(coverPath); err == nil {
			err = os.Remove(coverPath)
			if err != nil {
				return err
			}
		}
	}
	err = database.Instance.Unscoped().Where("id = ?", id).Delete(&database.File{}).Error
	return err
}

func GetFileById(id uint) (*database.File, error) {
	var file database.File
	err := database.Instance.Find(&file, id).Error
	return &file, err
}

func GetFileByPath(path string) (*database.File, error) {
	file := database.File{}
	err := database.Instance.Where("path = ?", path).First(&file).Error
	return &file, err
}

func NewFileTranscodeTask(id uint, format string, codec string) error {
	file, err := GetFileById(id)
	if err != nil {
		return err
	}
	oid := xid.New().String()
	outputFilename := filepath.Base(file.Path)
	ext := filepath.Ext(file.Path)
	outputFilename = strings.TrimSuffix(outputFilename, ext)
	outputFilename += fmt.Sprintf("_%s_%s.%s", oid, codec, format)
	output := filepath.Join(filepath.Dir(file.Path), outputFilename)
	response, err := youtrans.DefaultYouTransClient.CreateNewTask(&youtrans.CreateTaskRequestBody{
		Input:  file.Path,
		Output: output,
		Format: format,
		Codec:  codec,
	})
	if response != nil {
		logrus.Info(response.Id)
	}
	return err
}
func NewVideoFile(path string) (file database.File) {
	meta, metaerr := GetVideoFileMeta(path)
	if metaerr == nil {
		duration, err := strconv.ParseFloat(meta.GetFormat().GetDuration(), 10)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Duration = duration
		}

		size, err := strconv.ParseInt(meta.GetFormat().GetSize(), 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Size = size
		}

		bitrate, err := strconv.ParseInt(meta.GetFormat().GetBitRate(), 10, 64)
		if err != nil {
			VideoLogger.Error(err)
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
	}
	file.Path = path
	return
}
func CompleteTrans(tranTask youtrans.TaskResponse) error {
	original, err := GetFileByPath(tranTask.Input)
	if err != nil {
		return err
	}
	if original == nil {
		return errors.New("original trans file not found")
	}
	newFile := NewVideoFile(tranTask.Output)
	newFile.VideoId = original.VideoId

	// generate cover
	coverPath, err := GenerateVideoCover(newFile.Path)
	if err != nil {
		return err
	}
	newFile.Cover = filepath.Base(coverPath)
	err = database.Instance.Create(&newFile).Error
	if err != nil {
		return err
	}
	return nil
}

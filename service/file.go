package service

import (
	"fmt"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/youtrans"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
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

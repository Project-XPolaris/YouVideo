package service

import (
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"os"
	"path/filepath"
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

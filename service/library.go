package service

import (
	"errors"
	"github.com/projectxpolaris/youvideo/database"
	"gorm.io/gorm"
)

var (
	LibraryExistedError = errors.New("library existed")
)

func CreateLibrary(path string, name string) (*database.Library, error) {
	var recordCount int64
	err := database.Instance.Model(&database.Library{}).Where("path = ?", path).Count(&recordCount).Error
	if err != nil {
		return nil, err
	}
	if recordCount > 0 {
		return nil, LibraryExistedError
	}
	library := &database.Library{
		Path: path,
		Name: name,
	}
	err = database.Instance.Create(library).Error
	return library, err
}

func ScanLibrary(library *database.Library) error {
	return ScanVideo(library)
}

func ScanLibraryById(id uint) error {
	var library database.Library
	err := database.Instance.Find(&library, id).Error
	if err != nil {
		return err
	}
	CreateSyncLibraryTask(&library)
	return nil
}

type LibraryQueryOption struct {
	Page     int
	PageSize int
	Ids      []int64 `hsource:"query" hname:"id"`
}

func GetLibraryList(option LibraryQueryOption) (int64, []database.Library, error) {
	var result []database.Library
	var count int64
	queryBuilder := database.Instance.Model(&database.Library{})
	if len(option.Ids) > 0 {
		queryBuilder = queryBuilder.Where("id In ?", option.Ids)
	}
	err := queryBuilder.Limit(option.PageSize).Count(&count).Offset((option.Page - 1) * option.PageSize).Find(&result).Error
	return count, result, err
}

func RemoveLibraryById(id uint) error {
	var videos []database.Video
	err := database.Instance.
		Model(&database.Library{Model: gorm.Model{ID: id}}).
		Association("Videos").
		Find(&videos)
	if err != nil {
		return err
	}
	for _, video := range videos {
		err = DeleteVideoById(video.ID)
		if err != nil {
			return err
		}
	}
	return database.Instance.Unscoped().Delete(&database.Library{}, id).Error
}

func GetLibraryById(id uint) (*database.Library, error) {
	var library database.Library
	err := database.Instance.Find(&library, id).Error
	return &library, err
}

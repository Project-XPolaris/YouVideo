package service

import (
	"errors"
	"github.com/projectxpolaris/youvideo/database"
)

var (
	LibraryExistedError = errors.New("library existed")
)

func CreateLibrary(path string) (*database.Library, error) {
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
	err = ScanLibrary(&library)
	if err != nil {
		return err
	}
	return nil
}

type LibraryQueryOption struct {
	Page     int
	PageSize int
}

func GetLibraryList(option LibraryQueryOption) (int64, []database.Library, error) {
	var result []database.Library
	var count int64
	queryBuilder := database.Instance.Model(&database.Library{})
	err := queryBuilder.Limit(option.PageSize).Count(&count).Offset((option.Page - 1) * option.PageSize).Find(&result).Error
	return count, result, err
}

func RemoveLibraryById(id uint) error {
	err := database.Instance.
		Model(&database.Video{}).
		Unscoped().
		Where("library_id = ?", id).
		Delete(&database.Video{}).
		Error
	if err != nil {
		return err
	}
	return database.Instance.Unscoped().Delete(&database.Library{}, id).Error
}

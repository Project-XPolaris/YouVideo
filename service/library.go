package service

import (
	"errors"
	"github.com/projectxpolaris/youvideo/database"
	"gorm.io/gorm"
)

var (
	LibraryExistedError = errors.New("library existed")
	LibraryOwnerError   = errors.New("library not accessible")
)

func CreateLibrary(path string, name string, uid string, defaultVideoType string) (*database.Library, error) {
	var recordCount int64
	err := database.Instance.Model(&database.Library{}).Where("path = ?", path).Count(&recordCount).Error
	if err != nil {
		return nil, err
	}
	if recordCount > 0 {
		return nil, LibraryExistedError
	}
	user, err := GetUserById(uid)
	if err != nil {
		return nil, err
	}
	if len(defaultVideoType) == 0 {
		defaultVideoType = "video"
	}
	library := &database.Library{
		Path:             path,
		Name:             name,
		Users:            []*database.User{user},
		DefaultVideoType: defaultVideoType,
	}
	err = database.Instance.Create(library).Error
	return library, err
}

type LibraryQueryOption struct {
	Page     int
	PageSize int
	Ids      []int64 `hsource:"query" hname:"id"`
	Uid      string  `hsource:"param" hname:"uid"`
}

func GetLibraryList(option LibraryQueryOption) (int64, []database.Library, error) {
	var result []database.Library
	var count int64
	queryBuilder := database.Instance.Model(&database.Library{})
	if len(option.Ids) > 0 {
		queryBuilder = queryBuilder.Where("libraries.id In ?", option.Ids)
	}
	queryBuilder = queryBuilder.
		Joins("left join library_users on library_users.library_id = libraries.id").
		Joins("left join users on library_users.user_id = users.id").
		Where("users.uid in ?", []string{PublicUid, option.Uid})

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
	err = database.Instance.
		Model(&database.Library{Model: gorm.Model{ID: id}}).
		Association("Users").Clear()
	if err != nil {
		return err
	}
	return database.Instance.Unscoped().Delete(&database.Library{}, id).Error
}

func GetLibraryById(id uint, preloads ...string) (*database.Library, error) {
	var library database.Library
	query := database.Instance
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	err := query.First(&library, id).Error
	return &library, err
}

func CheckLibraryUidOwner(id uint, uid string) bool {
	var count int64
	database.Instance.
		Model(&database.Library{}).
		Joins("left join library_users on library_users.library_id = libraries.id").
		Joins("left join users on users.id = library_users.user_id").
		Where("users.uid in ?", []string{PublicUid, uid}).
		Where("libraries.id = ?", id).
		Count(&count)
	return count != 0
}

func CheckLibraryPathExist(path string) bool {
	var count int64
	database.Instance.
		Model(&database.Library{}).
		Where("path = ?", path).
		Count(&count)
	return count != 0
}

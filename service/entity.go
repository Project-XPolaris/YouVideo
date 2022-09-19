package service

import (
	"context"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
	"gorm.io/gorm"
)

func CreateEntity(name string, libraryId uint) (*database.Entity, error) {
	newData := database.Entity{
		Name:      name,
		LibraryId: libraryId,
	}
	err := database.Instance.Create(&newData).Error
	if err != nil {
		return nil, err
	}
	return &newData, nil
}

type EntityQueryBuilder struct {
	Id           uint   `hsource:"query" hname:"id"`
	Search       string `hsource:"query" hname:"search"`
	Page         int    `hsource:"param" hname:"page"`
	PageSize     int    `hsource:"param" hname:"pageSize"`
	Name         string `hsource:"query" hname:"name"`
	ReleaseStart string `hsource:"query" hname:"releaseStart"`
	ReleaseEnd   string `hsource:"query" hname:"releaseEnd"`
	Order        string `hsource:"query" hname:"order"`
}

func (e *EntityQueryBuilder) Query() ([]*database.Entity, int64, error) {
	var entities []*database.Entity
	var count int64
	query := database.Instance.Model(&database.Entity{})
	if e.Page == 0 {
		e.Page = 1
	}
	if e.PageSize == 0 {
		e.PageSize = 10
	}
	if e.Search != "" {
		query = query.Where("name LIKE ?", "%"+e.Search+"%")
	}
	if e.Name != "" {
		query = query.Where("name = ?", e.Name)
	}
	if e.Id != 0 {
		query = query.Where("id = ?", e.Id)
	}
	err := query.
		Preload("Videos").
		Preload("Videos.Files").
		Preload("Videos.Infos").
		Offset((e.Page - 1) * e.PageSize).
		Limit(e.PageSize).
		Find(&entities).Offset(-1).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	return entities, count, nil
}

func AddVideoToEntity(videoIds []uint, entityId uint) error {
	videoToAdd := make([]database.Video, 0)
	for _, videoId := range videoIds {
		videoToAdd = append(videoToAdd, database.Video{Model: gorm.Model{ID: videoId}})
	}

	err := database.Instance.Model(&database.Entity{Model: gorm.Model{ID: entityId}}).
		Association("Videos").
		Append(videoToAdd)
	if err != nil {
		return err
	}
	return nil
}

func GetOrCreateEntity(name string, libraryId uint) (*database.Entity, error) {
	var entity database.Entity
	err := database.Instance.Where("name = ?", name).Where("library_id = ?", libraryId).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return CreateEntity(name, libraryId)
		}
		return nil, err
	}
	return &entity, nil
}

func GetEntityById(id uint) (*database.Entity, error) {
	var entity database.Entity
	err := database.Instance.Where("id = ?", id).
		Preload("Videos").
		Preload("Videos.Files").
		First(&entity).
		Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func UpdateEntityById(id uint, updateData map[string]interface{}) (*database.Entity, error) {
	var entity database.Entity
	err := database.Instance.Where("id = ?", id).First(&entity).Error
	if coverUrl, isExist := updateData["coverUrl"]; isExist {
		sotreKey, err := DownloadEntityCover(coverUrl.(string))
		if err != nil {
			return nil, err
		}
		storage := plugin.GetDefaultStorage()
		reader, err := storage.Get(context.Background(), plugin.GetDefaultBucket(), "entity/"+sotreKey)
		if err != nil {
			return nil, err
		}
		width, height, err := util.GetImageSize(reader)
		if err != nil {
			return nil, err
		}
		updateData["cover_width"] = width
		updateData["cover_height"] = height
		updateData["cover"] = sotreKey
		delete(updateData, "coverUrl")
	}
	err = database.Instance.Model(entity).Updates(updateData).Error
	if err != nil {
		return nil, err
	}
	err = database.Instance.Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

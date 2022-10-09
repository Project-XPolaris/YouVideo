package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
	"gorm.io/gorm"
	"os"
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
	err = database.Instance.Find(&newData, newData.ID).Error
	return &newData, err
}

type EntityQueryBuilder struct {
	Id           uint     `hsource:"query" hname:"id"`
	Search       string   `hsource:"query" hname:"search"`
	Random       string   `hsource:"query" hname:"random"`
	Page         int      `hsource:"param" hname:"page"`
	PageSize     int      `hsource:"param" hname:"pageSize"`
	Name         string   `hsource:"query" hname:"name"`
	LibraryId    int      `hsource:"query" hname:"library"`
	ReleaseStart string   `hsource:"query" hname:"releaseStart"`
	ReleaseEnd   string   `hsource:"query" hname:"releaseEnd"`
	Orders       []string `hsource:"query" hname:"order"`
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
	if e.LibraryId != 0 {
		query = query.Where("library_id = ?", e.LibraryId)
	}
	if len(e.Random) > 0 {
		if database.Instance.Dialector.Name() == "sqlite" {
			query = query.Order("random()")
		} else {
			query = query.Order("RAND()")
		}
	} else {
		for _, order := range e.Orders {
			query = query.Order(fmt.Sprintf("entities.%s", order))
		}
	}
	err := query.
		Preload("Videos").
		Preload("Videos.Files").
		Preload("Videos.Infos").
		Preload("Tags").
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

func GetOrCreateEntityWithDirPath(name string, libraryId uint, dirPath string) (bool, *database.Entity, error) {
	var entity database.Entity
	err := database.Instance.
		Where("directory_path = ?", dirPath).
		Where("library_id = ?", libraryId).
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			entity := &database.Entity{
				Name:          name,
				DirectoryPath: dirPath,
				LibraryId:     libraryId,
			}
			err = database.Instance.Create(entity).Error
			return true, entity, err
		}
		return false, nil, err
	}
	return false, &entity, nil
}

func GetEntityById(id uint) (*database.Entity, error) {
	var entity database.Entity
	err := database.Instance.Where("id = ?", id).
		Preload("Videos").
		Preload("Videos.Files").
		Preload("Tags").
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

type EntityMetaFile struct {
	BangumiId string `json:"bangumiId"`
}

func ParseEntityMetaFile(entity *database.Entity, metaPath string) error {
	logScope := plugin.DefaultYouLogPlugin.Logger.NewScope("entity")
	logScope.Info("parse meta file ", "path=", metaPath)
	rawContent, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}
	meta := &EntityMetaFile{}
	err = json.Unmarshal(rawContent, meta)
	if err != nil {
		return err
	}
	if len(meta.BangumiId) > 0 {
		logScope.Info("parse bangumi id ", meta.BangumiId)
		if bangumiInfoSource != nil {
			return bangumiInfoSource.ApplyEntityById(entity, meta.BangumiId)
		}
	}
	return nil
}

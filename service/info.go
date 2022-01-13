package service

import (
	"github.com/projectxpolaris/youvideo/database"
	"gorm.io/gorm"
)

func AddVideoInfoItem(videoId uint, key string, value string) (*database.VideoMetaItem, error) {
	meta := &database.VideoMetaItem{
		Key:   key,
		Value: value,
	}
	err := database.Instance.
		Model(&database.VideoMetaItem{}).
		Where("key = ?", meta.Key).
		Where("value = ?", meta.Value).
		First(meta).Error
	if err == gorm.ErrRecordNotFound {
		err := database.Instance.Create(meta).Error
		if err != nil {
			return nil, err
		}
	}
	err = database.Instance.Model(meta).
		Association("Videos").
		Append(&database.Video{Model: gorm.Model{ID: videoId}})
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func RemoveInfoItem(id uint) error {
	err := database.Instance.Unscoped().Model(database.VideoMetaItem{}).Delete(id, "id").Error
	if err != nil {
		return err
	}
	return nil
}

type InfoQueryBuilder struct {
	Key      string `hsource:"query" hname:"key"`
	Value    string `hsource:"query" hname:"value"`
	Dist     string `hsource:"query" hname:"dist"`
	Page     int    `hsource:"query" hname:"page"`
	PageSize int    `hsource:"query" hname:"pageSize"`
}

func (e *InfoQueryBuilder) Query() ([]*database.VideoMetaItem, int64, error) {
	var entities []*database.VideoMetaItem
	var count int64
	query := database.Instance.Model(&database.VideoMetaItem{})
	if e.Page == 0 {
		e.Page = 1
	}
	if e.PageSize == 0 {
		e.PageSize = 10
	}
	if e.Dist != "" {
		query = query.Distinct("key")
	}
	if e.Key != "" {
		query = query.Where("key = ?", e.Key)
	}
	if e.Value != "" {
		query = query.Where("value = ?", e.Value)
	}
	err := query.
		Offset((e.Page - 1) * e.PageSize).
		Limit(e.PageSize).
		Find(&entities).Offset(-1).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	return entities, count, nil
}

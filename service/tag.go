package service

import (
	"github.com/allentom/haruka/gormh"
	"github.com/projectxpolaris/youvideo/database"
)

func GetTagByName(name string) (*database.Tag, error) {
	var tag database.Tag
	return &tag, database.Instance.First(&tag, "name = ?", name).Error
}

type TagQueryBuilder struct {
	gormh.DefaultPageFilter
}

func (t *TagQueryBuilder) ReadModels() (int64, interface{}, error) {
	models := make([]*database.Tag, 0)
	var count int64
	err := database.Instance.Model(&database.Tag{}).Limit(t.GetLimit()).Offset(t.GetOffset()).Find(&models).Count(&count).Error
	return count, models, err
}

func AddVideosToTag(modelId uint, ids ...interface{}) error {
	err := database.Instance.Model(&database.Tag{}).Where("id = ?", modelId).Association("Videos").Append(ids...)
	if err != nil {
		return err
	}
	return nil
}

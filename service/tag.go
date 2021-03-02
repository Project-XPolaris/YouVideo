package service

import (
	"github.com/allentom/haruka/gormh"
	"github.com/projectxpolaris/youvideo/database"
	"gorm.io/gorm"
)

func GetTagByName(name string) (*database.Tag, error) {
	var tag database.Tag
	return &tag, database.Instance.First(&tag, "name = ?", name).Error
}

type TagQueryBuilder struct {
	gormh.DefaultPageFilter
	TagVideoIdsQueryFilter
}

func (t *TagQueryBuilder) InVideoIds(ids ...interface{}) {
	if t.videoIds == nil {
		t.videoIds = []interface{}{}
	}
	t.videoIds = append(t.videoIds, ids...)
}

type TagVideoIdsQueryFilter struct {
	videoIds []interface{}
}

func (f TagVideoIdsQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.videoIds != nil && len(f.videoIds) > 0 {
		return db.Joins("left join video_tags on video_tags.tag_id = tags.id").Where("video_tags.video_id In ?", f.videoIds)
	}
	return db
}

func (t *TagQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = gormh.ApplyFilters(t, query)
	models := make([]*database.Tag, 0)
	var count int64
	err := query.Model(&database.Tag{}).Limit(t.GetLimit()).Offset(t.GetOffset()).Find(&models).Count(&count).Error
	return count, models, err
}

func AddVideosToTag(modelId uint, ids ...uint) error {
	appendVideo := make([]database.Video, 0)
	for _, id := range ids {
		appendVideo = append(appendVideo, database.Video{
			Model: gorm.Model{
				ID: id,
			},
		})
	}
	err := database.Instance.Model(&database.Tag{Model: gorm.Model{ID: modelId}}).Association("Videos").Append(appendVideo)
	if err != nil {
		return err
	}
	return nil
}

func RemoveVideosFromTag(modelId uint, ids ...uint) error {
	appendVideo := make([]database.Video, 0)
	for _, id := range ids {
		appendVideo = append(appendVideo, database.Video{
			Model: gorm.Model{
				ID: id,
			},
		})
	}
	err := database.Instance.Model(&database.Tag{Model: gorm.Model{ID: modelId}}).Association("Videos").Delete(appendVideo)
	if err != nil {
		return err
	}
	return nil
}

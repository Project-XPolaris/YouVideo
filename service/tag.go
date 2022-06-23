package service

import (
	"github.com/allentom/haruka/gormh"
	"github.com/projectxpolaris/youvideo/database"
	"gorm.io/gorm"
)

func GetTagByName(name string, uid string) (*database.Tag, error) {
	var tag database.Tag
	return &tag, database.Instance.
		Joins("left join tag_users on tag_users.tag_id = tags.id").
		Joins("left join users on tag_users.user_id = users.id").
		Where("users.uid in ?", []string{uid, PublicUid}).
		First(&tag, "name = ?", name).
		Error
}
func GetTagByID(id uint, uid string) (*database.Tag, error) {
	var tag database.Tag
	return &tag, database.Instance.
		Joins("left join tag_users on tag_users.tag_id = tags.id").
		Joins("left join users on tag_users.user_id = users.id").
		Where("users.uid in ?", []string{uid, PublicUid}).
		First(&tag, "tags.id = ?", id).
		Error
}

type TagQueryBuilder struct {
	Page       int    `hsource:"param" hname:"page"`
	PageSize   int    `hsource:"param" hname:"pageSize"`
	Video      []int  `hsource:"query" hname:"video"`
	SearchName string `hsource:"query" hname:"search"`
	Ids        []uint `hsource:"query" hname:"id"`
	Uid        string `hsource:"param" hname:"uid"`
}

func (t *TagQueryBuilder) Query() (int64, []*database.Tag, error) {
	query := database.Instance
	query = gormh.ApplyFilters(t, query)
	if len(t.SearchName) > 0 {
		query = query.Where("name like ?", "%"+t.SearchName+"%")
	}
	if len(t.Ids) > 0 {
		query = query.Where("id IN ?", t.Ids)
	}
	if len(t.Video) > 0 {
		query = query.Joins("left join video_tags on video_tags.tag_id = tags.id").
			Where("video_tags.video_id In ?", t.Video)
	}
	query = query.Joins("left join tag_users on tag_users.tag_id = tags.id").
		Joins("left join users on tag_users.user_id = users.id").
		Where("users.uid in ?", []string{t.Uid, PublicUid})
	models := make([]*database.Tag, 0)
	var count int64
	err := query.Model(&database.Tag{}).Limit(t.PageSize).Offset(t.PageSize * (t.Page - 1)).Find(&models).Count(&count).Error
	return count, models, err
}
func AddOrCreateTagFromVideo(tagName []string, uid string, ids ...uint) error {
	for _, name := range tagName {
		var tag database.Tag
		err := database.Instance.Where(&database.Tag{Name: name}).FirstOrCreate(&tag).Error
		if err != nil {
			return err
		}
		user, err := GetUserById(uid)
		if err != nil {
			return err
		}
		err = database.Instance.Model(&tag).Association("Users").Append(user)
		if err != nil {
			return err
		}
		videos := make([]interface{}, 0)
		for _, id := range ids {
			videos = append(videos, &database.Video{
				Model: gorm.Model{
					ID: id,
				},
			})
		}
		err = database.Instance.Model(&tag).Association("Videos").Append(videos...)
		if err != nil {
			return err
		}
	}
	return nil
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

func CreateTag(name string, uid string) (*database.Tag, error) {
	user, err := GetUserById(uid)
	if err != nil {
		return nil, err
	}
	tag := &database.Tag{
		Name: name,
		Users: []*database.User{
			user,
		},
	}
	err = database.Instance.Create(&tag).Error
	return tag, err
}

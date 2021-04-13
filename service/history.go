package service

import (
	"errors"
	"github.com/projectxpolaris/youvideo/database"
	"gorm.io/gorm"
	"time"
)

// create new play history
func CreateHistory(videoId uint, uid string) error {
	if uid == PublicUid {
		return nil
	}
	var user database.User
	err := database.Instance.First(&user, "uid = ?", uid).Error
	if err != nil {
		return err
	}
	history := database.History{UserID: user.ID, VideoID: videoId}
	err = database.Instance.First(&history).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		err = database.Instance.Model(&history).Where("id  = ?", history.ID).Update("updated_at", time.Now()).Error
		if err != nil {
			return err
		}
		return nil
	}
	err = database.Instance.Save(&database.History{
		UserID:  user.ID,
		VideoID: videoId,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

type HistoryQueryOption struct {
	Page     int    `hsource:"param" hname:"page"`
	PageSize int    `hsource:"param" hname:"pageSize"`
	Uid      string `hsource:"param" hname:"uid"`
}

func GetHistoryList(option HistoryQueryOption) (int64, []*database.History, error) {
	var result []*database.History
	var count int64
	queryBuilder := database.Instance.Model(&database.History{})
	queryBuilder = queryBuilder.
		Preload("User").Preload("Video").Preload("Video.Files").
		Joins("left join users on users.id = histories.user_id").
		Where("users.uid = ?", option.Uid)
	err := queryBuilder.Limit(option.PageSize).Count(&count).Offset((option.Page - 1) * option.PageSize).Find(&result).Error
	return count, result, err
}

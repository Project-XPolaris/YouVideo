package database

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Name   string
	Videos []*Video `gorm:"many2many:video_tags;"`
	Users  []*User  `gorm:"many2many:tag_users;"`
}

func (t *Tag) Save() error {
	return Instance.Create(t).Error
}

func (t *Tag) DeleteById(id uint) error {
	return Instance.Unscoped().Delete(&Tag{}, id).Error
}

func (t *Tag) UpdateById(id uint, values map[string]interface{}) (interface{}, error) {
	err := Instance.Model(&Tag{}).Where("id = ?", id).Updates(values).Error
	if err != nil {
		return nil, err
	}
	tag := &Tag{
		Model: gorm.Model{ID: id},
	}
	return tag, Instance.Find(tag).Error
}

package service

import "github.com/projectxpolaris/youvideo/database"

type FolderQueryBuilder struct {
	Page      int    `hsource:"param" hname:"page"`
	PageSize  int    `hsource:"param" hname:"pageSize"`
	LibraryId string `hsource:"query" hname:"library"`
}

func (f *FolderQueryBuilder) Read() (int64, []*database.Folder, error) {
	query := database.Instance.Where("library_id = ?", f.LibraryId)
	var count int64
	var result []*database.Folder
	err := query.Model(&database.Folder{}).Preload("Videos").Preload("Videos.Files").Limit(f.PageSize).Count(&count).Offset((f.Page - 1) * f.PageSize).Find(&result).Error
	return count, result, err
}

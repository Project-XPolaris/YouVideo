package service

import (
	"errors"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
)

func SearchWithMeiliSearch(searchKey string, uid string) (*SearchHit, error) {
	if !DefaultMeilisearchEngine.Enable {
		return nil, errors.New("meilisearch is not enabled")
	}
	accessibleDatabase, err := GetUserAccessibleDatabase(uid)
	if err != nil {
		return nil, err
	}
	libraryIds := make([]uint, 0)
	for _, library := range accessibleDatabase {
		libraryIds = append(libraryIds, library.ID)
	}
	result, err := DefaultMeilisearchEngine.Search(searchKey, libraryIds)
	if err != nil {
		return nil, err
	}
	videoIdToSearch := make([]uint, 0)
	for _, videoIndex := range result.Videos {
		video := videoIndex.(map[string]interface{})
		id := video["id"].(float64)
		videoId := uint(id)
		videoIdToSearch = append(videoIdToSearch, videoId)
	}
	var videos []*database.Video
	err = database.Instance.Where("id in (?)", videoIdToSearch).
		Preload("Files").Preload("Infos").Preload("Files.Subtitles").
		Find(&videos).Error
	if err != nil {
		return nil, err
	}
	entityIdToSearch := make([]uint, 0)
	for _, entityIndex := range result.Entities {
		entity := entityIndex.(map[string]interface{})
		id := entity["id"].(float64)
		entityId := uint(id)
		entityIdToSearch = append(entityIdToSearch, entityId)
	}
	var entities []*database.Entity
	err = database.Instance.Where("id in (?)", entityIdToSearch).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return &SearchHit{
		Videos:   videos,
		Entities: entities,
	}, nil
}
func SearchWithDatabase(searchKey string, uid string) (*SearchHit, error) {
	accessibleDatabase, err := GetUserAccessibleDatabase(uid)
	if err != nil {
		return nil, err
	}
	libraryIds := make([]uint, 0)
	for _, library := range accessibleDatabase {
		libraryIds = append(libraryIds, library.ID)
	}
	var videos []*database.Video
	err = database.Instance.Where("library_id in (?)", libraryIds).
		Where("name like ?", "%"+searchKey+"%").
		Preload("Files").Preload("Infos").Preload("Files.Subtitles").
		Find(&videos).Error
	if err != nil {
		return nil, err
	}
	var entities []*database.Entity
	err = database.Instance.Where("library_id in (?)", libraryIds).
		Where("name like ?", "%"+searchKey+"%").
		Or("summary like ?", "%"+searchKey+"%").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return &SearchHit{
		Videos:   videos,
		Entities: entities,
	}, nil

}

func SearchData(searchKey string, uid string) (*SearchHit, error) {
	switch config.Instance.SearchEngine {
	case "meilisearch":
		return SearchWithMeiliSearch(searchKey, uid)
	default:
		return SearchWithDatabase(searchKey, uid)
	}
}

package service

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
)

var DefaultMeilisearchEngine = &MeilisearchEngine{}

type MeilisearchEngine struct {
	client *meilisearch.Client
	logger *youlog.Scope
	Enable bool
}

type SearchIndex struct {
	Name             string
	SearchableFields []string
	FilterableFields []string
	PrimaryKey       string
}

var VideosIndex = SearchIndex{
	Name:             "videos",
	SearchableFields: []string{"name"},
	FilterableFields: []string{"libraryId"},
	PrimaryKey:       "id",
}
var EntityIndex = SearchIndex{
	Name:             "entity",
	SearchableFields: []string{"name", "summary"},
	FilterableFields: []string{"libraryId"},
	PrimaryKey:       "id",
}
var Indexes = []SearchIndex{
	VideosIndex, EntityIndex,
}

func InitMeiliSearch() error {
	logger := plugin.DefaultYouLogPlugin.Logger.NewScope("meilisearch")
	DefaultMeilisearchEngine.logger = logger
	logger.Info("init meilisearch engine")
	if plugin.DefaultMeiliSearchPlugin.Client == nil {
		logger.Info("meilisearch client not enable")
		DefaultMeilisearchEngine.Enable = false
		return nil
	}
	DefaultMeilisearchEngine.Enable = true
	client := plugin.DefaultMeiliSearchPlugin.Client
	DefaultMeilisearchEngine.client = client
	logger.Info("check videos index")
	queryIndexResponse, err := client.GetIndexes(&meilisearch.IndexesQuery{
		Limit:  0,
		Offset: 0,
	})
	if err != nil {
		return err
	}
	missIndexs := make([]SearchIndex, 0)
	for _, index := range Indexes {
		isExist := false
		for _, existIndex := range queryIndexResponse.Results {
			if existIndex.UID == index.Name {
				isExist = true
				break
			}
		}
		if !isExist {
			missIndexs = append(missIndexs, index)
		}
	}
	for _, missIndex := range missIndexs {
		logger.Info(fmt.Sprintf("create index %s", missIndex))
		_, err = client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        missIndex.Name,
			PrimaryKey: missIndex.PrimaryKey,
		})
		if err != nil {
			return err
		}
		_, err = client.Index(missIndex.Name).UpdateSearchableAttributes(&missIndex.SearchableFields)
		if err != nil {
			return err
		}
		_, err = client.Index(missIndex.Name).UpdateFilterableAttributes(&missIndex.FilterableFields)
		if err != nil {
			return err
		}
	}
	//client.DeleteIndex("videos")
	//client.DeleteIndex("entity")
	return nil
}

type VideoDoc struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	LibraryId uint   `json:"libraryId"`
}

type EntityDoc struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	LibraryId uint   `json:"libraryId"`
	Summary   string `json:"summary"`
}

func (e *MeilisearchEngine) Sync(LibraryId uint) error {
	e.logger.Info(fmt.Sprintf("start sync library %d", LibraryId))
	var videos []database.Video
	err := database.Instance.Where("library_id = ?", LibraryId).Find(&videos).Error
	if err != nil {
		return err
	}
	// sync videos index
	e.logger.Info(fmt.Sprintf("total %d videos to sync", len(videos)))
	videosIndex := e.client.Index("videos")
	existedIndex, err := videosIndex.Search("", &meilisearch.SearchRequest{
		Filter: [][]string{
			{fmt.Sprintf("libraryId = %d", LibraryId)},
		},
	})
	if err != nil {
		return err
	}
	pageSize := existedIndex.EstimatedTotalHits
	existedIndex, err = videosIndex.Search("", &meilisearch.SearchRequest{
		Filter: [][]string{
			{fmt.Sprintf("libraryId = %d", LibraryId)},
		},
		Limit:  pageSize,
		Offset: 0,
	})

	if err != nil {
		return err
	}
	videoIdsToDelete := make([]string, 0)
	for _, hit := range existedIndex.Hits {
		id := hit.(map[string]interface{})["id"].(float64)
		isExist := false
		for _, video := range videos {
			if video.ID == uint(id) {
				isExist = true
				break
			}
		}
		if !isExist {
			videoIdsToDelete = append(videoIdsToDelete, fmt.Sprintf("%f", id))
		}
	}
	videosIndex.DeleteDocuments(videoIdsToDelete)
	docs := make([]VideoDoc, 0)
	for _, video := range videos {
		docs = append(docs, VideoDoc{
			Id:        video.ID,
			Name:      video.Name,
			LibraryId: video.LibraryId,
		})
	}
	_, err = videosIndex.AddDocuments(docs, "id")
	if err != nil {
		return err
	}
	// sync entity
	var entities []database.Entity
	err = database.Instance.Where("library_id = ?", LibraryId).Find(&entities).Error
	if err != nil {
		return err
	}
	e.logger.Info(fmt.Sprintf("total %d entities to sync", len(entities)))
	entityIndex := e.client.Index("entity")
	existedIndex, err = entityIndex.Search("", &meilisearch.SearchRequest{
		Filter: [][]string{
			{fmt.Sprintf("libraryId = %d", LibraryId)},
		},
	})
	if err != nil {
		return err
	}
	pageSize = existedIndex.EstimatedTotalHits
	existedIndex, err = entityIndex.Search("", &meilisearch.SearchRequest{
		Filter: [][]string{
			{fmt.Sprintf("libraryId = %d", LibraryId)},
		},
		Limit:  pageSize,
		Offset: 0,
	})
	if err != nil {
		return err
	}
	entityIdsToDelete := make([]string, 0)
	for _, hit := range existedIndex.Hits {
		id := hit.(map[string]interface{})["id"].(float64)
		isExist := false
		for _, entity := range entities {
			if entity.ID == uint(id) {
				isExist = true
				break
			}
		}
		if !isExist {
			entityIdsToDelete = append(entityIdsToDelete, fmt.Sprintf("%f", id))
		}
	}
	entityIndex.DeleteDocuments(entityIdsToDelete)
	entityDocs := make([]EntityDoc, 0)
	for _, entity := range entities {
		entityDocs = append(entityDocs, EntityDoc{
			Id:        entity.ID,
			Name:      entity.Name,
			LibraryId: entity.LibraryId,
			Summary:   entity.Summary,
		})
	}
	_, err = entityIndex.AddDocuments(entityDocs, "id")
	if err != nil {
		return err
	}
	return nil
}

type SearchResult struct {
	Videos   []interface{} `json:"videos"`
	Entities []interface{} `json:"entities"`
}

func (e *MeilisearchEngine) Search(searchKey string, libraryIds []uint) (*SearchResult, error) {
	libraryIdFilters := make([]string, 0)
	for _, libraryId := range libraryIds {
		libraryIdFilters = append(libraryIdFilters, fmt.Sprintf("libraryId = %d", libraryId))
	}
	result := &SearchResult{}
	videoSearchResult, err := e.client.Index("videos").Search(searchKey, &meilisearch.SearchRequest{
		Filter: [][]string{
			libraryIdFilters,
		},
	})
	if err != nil {
		return nil, err
	}
	result.Videos = videoSearchResult.Hits
	entitySearchResult, err := e.client.Index("entity").Search(searchKey, &meilisearch.SearchRequest{
		Filter: [][]string{
			libraryIdFilters,
		},
	})
	if err != nil {
		return nil, err
	}
	result.Entities = entitySearchResult.Hits
	return result, nil
}

type SearchHit struct {
	Videos   []*database.Video  `json:"videos"`
	Entities []*database.Entity `json:"entities"`
}

func (e *MeilisearchEngine) GetIndexByLibraryId(libraryId uint, index string) (*meilisearch.SearchResponse, error) {
	videosIndex := e.client.Index(index)
	existedIndex, err := videosIndex.Search("", &meilisearch.SearchRequest{
		Filter: [][]string{
			{fmt.Sprintf("libraryId = %d", libraryId)},
		},
	})
	if err != nil {
		return nil, err
	}
	pageSize := existedIndex.EstimatedTotalHits
	existedIndex, err = videosIndex.Search("", &meilisearch.SearchRequest{
		Filter: [][]string{
			{fmt.Sprintf("libraryId = %d", libraryId)},
		},
		Limit:  pageSize,
		Offset: 0,
	})
	return existedIndex, nil
}
func (e *MeilisearchEngine) RemoveIndexByLibrary(libraryId uint, index string) error {
	videosIndex := e.client.Index(index)
	allVideosIndex, err := e.GetIndexByLibraryId(libraryId, index)
	if err != nil {
		return err
	}
	videoIdsToDelete := make([]string, 0)
	for _, hit := range allVideosIndex.Hits {
		id := hit.(map[string]interface{})["id"].(float64)
		videoIdsToDelete = append(videoIdsToDelete, fmt.Sprintf("%d", int64(id)))
	}
	_, err = videosIndex.DeleteDocuments(videoIdsToDelete)
	return err
}

func (e *MeilisearchEngine) DeleteAllIndexByLibrary(libraryId uint) error {
	err := e.RemoveIndexByLibrary(libraryId, "videos")
	if err != nil {
		return err
	}
	err = e.RemoveIndexByLibrary(libraryId, "entity")
	if err != nil {
		return err
	}
	return nil
}

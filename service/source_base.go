package service

import (
	"context"
	"errors"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"net/http"
	"path/filepath"
)

type SearchMovieResult struct {
	Name    string
	Cover   string
	Summary string
}
type SearchTVResult struct {
	Name    string
	Cover   string
	Summary string
}

type InfoSource interface {
	SearchMovie(query string) (*SearchMovieResult, error)
	SearchTv(query string) (*SearchTVResult, error)
	MatchEntity(entity *database.Entity) error
	DownloadCover(url string) (string, error)
	SearchTvList(query string) ([]*SearchTVResult, error)
	SearchMovieList(query string) ([]*SearchMovieResult, error)
}

func GetInfoSource(name string) InfoSource {
	if len(name) == 0 {
		return nil
	}
	switch name {
	case "tmdb":
		return tmdbSource
	case "bangumi":
		return bangumiInfoSource
	}
	return nil
}

func DownloadEntityCover(url string) (string, error) {
	if len(url) == 0 {
		return "", errors.New("url is empty")
	}
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", errors.New("received non 200 response code")
	}
	baseFileName := filepath.Base(url)
	key := "entity/" + baseFileName
	storage := plugin.GetDefaultStorage()
	err = storage.Upload(context.Background(), response.Body, plugin.GetDefaultBucket(), key)
	if err != nil {
		return "", err
	}
	return baseFileName, nil
}

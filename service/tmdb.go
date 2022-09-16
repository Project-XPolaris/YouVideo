package service

import (
	"context"
	"errors"
	"fmt"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
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

var tmdbClient *tmdb.Client

func InitTMDB() {
	var err error
	tmdbClient, err = tmdb.Init("7a89accf9a240cba6a2e56c0ff91cb00")
	if err != nil {
		fmt.Println(err)
	}
	proxyUrl, err := url.Parse("http://localhost:7890")
	customClient := http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 15 * time.Second,
			Proxy:           http.ProxyURL(proxyUrl),
		},
	}
	tmdbClient.SetClientConfig(customClient)

}

func SearchMovie(query string) (*SearchMovieResult, error) {
	movie, err := tmdbClient.GetSearchMovies(query, nil)
	if err != nil {
		return nil, err
	}
	if len(movie.SearchMoviesResults.Results) > 0 {
		searchResult := &SearchMovieResult{
			Name:    movie.SearchMoviesResults.Results[0].Title,
			Cover:   movie.SearchMoviesResults.Results[0].PosterPath,
			Summary: movie.SearchMoviesResults.Results[0].Overview,
		}
		return searchResult, nil
	}
	return nil, nil
}
func SearchTv(query string) (*SearchTVResult, error) {
	tvList, err := tmdbClient.GetSearchTVShow(query, nil)
	if err != nil {
		return nil, err
	}
	if len(tvList.SearchTVShowsResults.Results) > 0 {
		tv := tvList.SearchTVShowsResults.Results[0]
		result := &SearchTVResult{
			Name:    tv.Name,
			Cover:   tv.PosterPath,
			Summary: tv.Overview,
		}
		return result, nil
	}
	return nil, nil
}
func DownloadCover(url string) (string, error) {
	prefix := "https://image.tmdb.org/t/p/w500"
	response, err := http.Get(prefix + url)
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
func MatchEntity(entity *database.Entity) error {
	switch entity.Template {
	case "film":
		movie, err := SearchMovie(entity.Name)
		if err != nil {
			return err
		}
		if movie != nil {
			entity.Name = movie.Name
			entity.Cover = movie.Cover
			entity.Summary = movie.Summary
		}
	case "tv":
		tv, err := SearchTv(entity.Name)
		if err != nil {
			return err
		}
		if tv != nil {
			entity.Name = tv.Name
			entity.Cover = tv.Cover
			entity.Summary = tv.Summary
		}

	}
	return nil
}

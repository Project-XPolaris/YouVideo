package service

import (
	"context"
	"errors"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

var tmdbSource *TMDBSource

type TMDBSource struct {
	tmdbClient *tmdb.Client
	logScope   *youlog.Scope
}

func InitTMDB() {
	var err error
	logScope := plugin.DefaultYouLogPlugin.Logger.NewScope("tmdb")
	tmdbConfig := config.Instance.TMdbConfig
	logScope.Info("init tmdb")
	if !tmdbConfig.Enable {
		logScope.Info("tmdb is disabled")
		return
	}
	if len(tmdbConfig.ApiKey) == 0 {
		logScope.Fatal("tmdb api key is empty")
		return
	}

	tmdbSource = &TMDBSource{
		logScope: logScope,
	}
	tmdbClient, err := tmdb.Init(tmdbConfig.ApiKey)
	if err != nil {
		logScope.Fatal(err)
		return
	}
	tmdbSource.tmdbClient = tmdbClient
	customClient := http.Client{
		Timeout: time.Second * 5,
	}
	clientTransport := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 15 * time.Second,
	}
	if len(tmdbConfig.Proxy) > 0 {
		proxyUrl, err := url.Parse(tmdbConfig.Proxy)
		if err != nil {
			logScope.Fatal(err)
			return
		}
		clientTransport.Proxy = http.ProxyURL(proxyUrl)
	}
	customClient.Transport = clientTransport
	tmdbClient.SetClientConfig(customClient)
}

func (s *TMDBSource) SearchMovie(query string) (*SearchMovieResult, error) {
	movie, err := s.tmdbClient.GetSearchMovies(query, nil)
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
func (s *TMDBSource) SearchTv(query string) (*SearchTVResult, error) {
	tvList, err := s.tmdbClient.GetSearchTVShow(query, nil)
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
func (s *TMDBSource) DownloadCover(url string) (string, error) {
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
func (s *TMDBSource) MatchEntity(entity *database.Entity) error {
	switch entity.Template {
	case "film":
		movie, err := s.SearchMovie(entity.Name)
		if err != nil {
			return err
		}
		if movie != nil {
			entity.Name = movie.Name
			entity.Cover = movie.Cover
			entity.Summary = movie.Summary
		}
	case "tv":
		tv, err := s.SearchTv(entity.Name)
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

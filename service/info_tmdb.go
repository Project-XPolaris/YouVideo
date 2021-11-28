package service

import (
	"fmt"
	tmdb "github.com/cyruzin/golang-tmdb"
	"strconv"
	"time"
)

type TMDBInformationProvider struct {
	Client    *tmdb.Client
	Available bool
}

func (p *TMDBInformationProvider) IsAvailable() bool {
	return p.Available
}
func (p *TMDBInformationProvider) Init() error {
	tmdbClient, err := tmdb.Init("7a89accf9a240cba6a2e56c0ff91cb00")
	if err != nil {
		return err
	}
	p.Client = tmdbClient
	p.Available = true
	return nil
}
func (p *TMDBInformationProvider) SearchVideo(keyword string) ([]SearchMovieResult, error) {
	result, err := p.Client.GetSearchMovies(keyword, map[string]string{})
	if err != nil {
		return nil, err
	}
	resultMovies := make([]SearchMovieResult, 0)
	timeformat := "2006-01-02"
	for _, searchResult := range result.SearchMoviesResults.Results {
		movie := SearchMovieResult{
			Id:      fmt.Sprintf("%d", searchResult.ID),
			Name:    searchResult.Title,
			Summary: searchResult.Overview,
		}
		if len(searchResult.PosterPath) > 0 {
			movie.Cover = fmt.Sprintf("https://image.tmdb.org/t/p/original%s", searchResult.PosterPath)
		}
		if len(searchResult.ReleaseDate) > 0 {
			movie.Release, _ = time.Parse(timeformat, searchResult.ReleaseDate)
		}
		resultMovies = append(resultMovies, movie)
	}
	return resultMovies, nil
}

func (p *TMDBInformationProvider) GetMovieDetail(Id string) (*MovieDetail, error) {
	numberId, err := strconv.Atoi(Id)
	if err != nil {
		return nil, err
	}
	detailResponse, err := p.Client.GetMovieDetails(numberId, map[string]string{})
	if err != nil {
		return nil, err
	}
	timeformat := "2006-01-02"
	detail := &MovieDetail{
		Id:      Id,
		Name:    detailResponse.Title,
		Cover:   fmt.Sprintf("https://image.tmdb.org/t/p/original%s", detailResponse.PosterPath),
		Summary: detailResponse.Overview,
		Release: time.Time{},
	}
	if len(detailResponse.ReleaseDate) > 0 {
		detail.Release, _ = time.Parse(timeformat, detailResponse.ReleaseDate)
	}
	return detail, nil
}

func (p *TMDBInformationProvider) GetMovieCredit(id string) ([]MovieCredit, error) {
	numberId, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	creditResponse, err := p.Client.GetMovieCredits(numberId, map[string]string{})
	if err != nil {
		return nil, err
	}
	creditList := make([]MovieCredit, 0)
	for _, s := range creditResponse.Cast {
		credit := MovieCredit{
			Id:        fmt.Sprintf("%d", s.ID),
			Name:      s.Name,
			Pic:       fmt.Sprintf("https://image.tmdb.org/t/p/original%s", s.ProfilePath),
			Character: s.Character,
		}
		creditList = append(creditList, credit)
	}
	return creditList, nil
}

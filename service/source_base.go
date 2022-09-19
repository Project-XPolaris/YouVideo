package service

import "github.com/projectxpolaris/youvideo/database"

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

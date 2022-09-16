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
}

func GetInfoSource() InfoSource {
	if tmdbSource != nil {
		return tmdbSource
	}
	if bangumiInfoSource != nil {
		return bangumiInfoSource
	}
	return nil
}

package service

import (
	"github.com/projectxpolaris/youvideo/database"
	"time"
)

type VideoInformationProvider interface {
	Init() error
	IsAvailable() bool
	SearchVideo(keyword string) ([]SearchMovieResult, error)
	GetMovieDetail(Id string) (*MovieDetail, error)
	GetMovieCredit(id string) ([]MovieCredit, error)
}

type SearchMovieResult struct {
	Id      string
	Name    string
	Cover   string
	Summary string
	Release time.Time
}

type MovieDetail struct {
	Id      string
	Name    string
	Cover   string
	Summary string
	Release time.Time
}
type MovieCredit struct {
	Id        string
	Name      string
	Pic       string
	Character string
}

func MatchVideoInformationById(id int) error {
	var video database.Video
	err := database.Instance.First(&video, id).Error
	if err != nil {
		return err
	}
	wrap := NewMatchVideoInformationWrap(&video)
	DefaultMovieInformationMatcher.In <- wrap
	select {
	case <-wrap.OnDone:
		return nil
	case err = <-wrap.OnError:
		return err
	}
}

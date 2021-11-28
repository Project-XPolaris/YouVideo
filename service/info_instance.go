package service

import (
	"context"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/youlog"
)

var DefaultMovieInformationMatcher = MovieInformationMatcher{
	Providers: []VideoInformationProvider{
		&TMDBInformationProvider{},
	},
}

type MovieInformationMatcher struct {
	In        chan *MatchVideoInformationWrap
	Providers []VideoInformationProvider
}
type MatchVideoInformationWrap struct {
	Video   *database.Video
	OnError chan error
	OnDone  chan struct{}
}

func (w *MatchVideoInformationWrap) Done() {
	if w.OnDone != nil {
		w.OnDone <- struct{}{}
	}
}
func (w *MatchVideoInformationWrap) Error(err error) {
	if w.OnError != nil {
		w.OnError <- err
	}
}
func NewMatchVideoInformationWrap(video *database.Video) *MatchVideoInformationWrap {
	return &MatchVideoInformationWrap{
		Video:   video,
		OnError: make(chan error),
		OnDone:  make(chan struct{}),
	}
}
func (m *MovieInformationMatcher) Run(ctx context.Context) error {
	m.In = make(chan *MatchVideoInformationWrap, 100)
	for _, infoProvider := range m.Providers {
		err := infoProvider.Init()
		if err != nil {
			return err
		}
	}
	logger := youlog.DefaultYouLogPlugin.Logger.NewScope("MovieInformationMatcher")
	go func() {
		for {
			var err error
			select {
			case <-ctx.Done():
				return
			case videoWrap := <-m.In:
				video := videoWrap.Video
				var searchMovieResult []SearchMovieResult
				var targetProvider VideoInformationProvider
				for _, provider := range m.Providers {
					if !provider.IsAvailable() {
						continue
					}
					searchMovieResult, err = provider.SearchVideo(video.Name)
					if err != nil {
						logger.Error(err.Error())
					}
					if searchMovieResult != nil {
						targetProvider = provider
						break
					}
				}
				// not search found
				if searchMovieResult == nil {
					videoWrap.Done()
					continue
				}

				var targetResult *SearchMovieResult
				for _, result := range searchMovieResult {
					distance := levenshtein.ComputeDistance(result.Name, video.Name)
					logger.Info(fmt.Sprintf("left = %s,right = %s,distance = %d", result.Name, video.Name, distance))
					if distance < 3 {
						targetResult = &result
						break
					}
				}
				// no match result
				if targetResult == nil {
					videoWrap.Done()
					continue
				}

				// get detail
				detail, err := targetProvider.GetMovieDetail(targetResult.Id)
				newInfo := database.MovieInformation{
					Title:   detail.Name,
					Cover:   detail.Cover,
					Release: detail.Release,
				}
				err = database.Instance.Save(&newInfo).Error
				if err != nil {
					logger.Error(err.Error())
					videoWrap.Error(err)
					continue
				}
				video.MovieInformation.ID = newInfo.ID
				err = database.Instance.Save(video).Error
				if err != nil {
					logger.Error(err.Error())
					videoWrap.Error(err)
					continue
				}
				// get cast
				castList, err := targetProvider.GetMovieCredit(detail.Id)
				if err != nil {
					logger.Error(err.Error())
					videoWrap.Error(err)
					continue
				}
				saveCastList := make([]database.MovieCredit, 0)
				for _, credit := range castList {
					saveCastList = append(saveCastList, database.MovieCredit{
						MovieInformationID: newInfo.ID,
						Name:               credit.Name,
						Pic:                credit.Pic,
						Character:          credit.Character,
					})
				}
				err = database.Instance.Create(&saveCastList).Error
				if err != nil {
					logger.Error(err.Error())
					videoWrap.Error(err)
					continue
				}
				videoWrap.Done()
			}
		}
	}()
	return nil
}

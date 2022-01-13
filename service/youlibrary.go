package service

import (
	"context"
	"errors"
	"github.com/project-xpolaris/youplustoolkit/youlibrary"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	youlog2 "github.com/projectxpolaris/youvideo/youlog"
)

var DefaultVideoInformationMatchService = VideoInformationMatchService{}

type VideoInformationMatchInput struct {
	Video   *database.Video
	OnDone  chan struct{}
	OnError chan error
}

func (i *VideoInformationMatchInput) Done() {
	i.OnDone <- struct{}{}
}
func (i *VideoInformationMatchInput) RaiseError(err error) {
	i.OnError <- err
}
func NewVideoInformationMatchInput(video *database.Video) *VideoInformationMatchInput {
	return &VideoInformationMatchInput{
		Video:   video,
		OnDone:  make(chan struct{}),
		OnError: make(chan error),
	}
}

type VideoInformationMatchService struct {
	In     chan *VideoInformationMatchInput
	Client *youlibrary.Client
	Logger *youlog.Scope
}

func (s *VideoInformationMatchService) Init() {
	s.In = make(chan *VideoInformationMatchInput, 1000)
	s.Client = youlibrary.NewYouLibraryClient()
	s.Client.Init(config.Instance.YouLibraryConfig.Url)
	s.Logger = youlog2.DefaultYouLogPlugin.Logger.NewScope("VideoInformationMatch")
	s.Logger.Info("init success")
}

func (s *VideoInformationMatchService) Run(context context.Context) {
	for {
		select {
		case input := <-s.In:
			video := input.Video
			response, err := s.Client.MatchVideoInfo(video.Name, video.Type)
			if err != nil {
				input.RaiseError(err)
				return
			}
			if !response.Success {
				input.RaiseError(errors.New("match error"))
				return
			}
			// link video and subject
			video.SubjectId = response.Data.Id
			err = database.Instance.Save(&video).Error
			if err != nil {
				input.RaiseError(err)
			}
		case <-context.Done():
			return
		}
	}
}

func MatchVideoInformationById(id int) error {
	var video database.Video
	err := database.Instance.First(&video, id).Error
	if err != nil {
		return err
	}
	input := NewVideoInformationMatchInput(&video)
	DefaultVideoInformationMatchService.In <- input
	select {
	case <-input.OnDone:
		return nil
	case err = <-input.OnError:
		return err
	}
}

func GetSubjectById(subjectId uint) (*youlibrary.GetSubjectResponse, error) {
	return DefaultVideoInformationMatchService.Client.GetSubjectById(subjectId)
}

package service

import (
	"context"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"net/http"
	"path/filepath"
)

type BangumiClient struct {
	client *resty.Client
}

type SearchObjectOption struct {
	Type          string
	ResponseGroup string
}
type SearchSubjectResult struct {
	Results int `json:"results"`
	List    []struct {
		ID         int    `json:"id"`
		URL        string `json:"url"`
		Type       int    `json:"type"`
		Name       string `json:"name"`
		NameCn     string `json:"name_cn"`
		Summary    string `json:"summary"`
		Eps        int    `json:"eps"`
		EpsCount   int    `json:"eps_count"`
		AirDate    string `json:"air_date"`
		AirWeekday int    `json:"air_weekday"`
		Rating     struct {
			Total int `json:"total"`
			Count struct {
				Num1  int `json:"1"`
				Num2  int `json:"2"`
				Num3  int `json:"3"`
				Num4  int `json:"4"`
				Num5  int `json:"5"`
				Num6  int `json:"6"`
				Num7  int `json:"7"`
				Num8  int `json:"8"`
				Num9  int `json:"9"`
				Num10 int `json:"10"`
			} `json:"count"`
			Score float64 `json:"score"`
		} `json:"rating,omitempty"`
		Rank   int `json:"rank,omitempty"`
		Images struct {
			Large  string `json:"large"`
			Common string `json:"common"`
			Medium string `json:"medium"`
			Small  string `json:"small"`
			Grid   string `json:"grid"`
		} `json:"images"`
		Collection struct {
			Wish    int `json:"wish"`
			Collect int `json:"collect"`
			Doing   int `json:"doing"`
			OnHold  int `json:"on_hold"`
			Dropped int `json:"dropped"`
		} `json:"collection"`
	} `json:"list"`
}

func NewBangumiClient() *BangumiClient {
	return &BangumiClient{
		client: resty.New(),
	}
}
func (c *BangumiClient) SearchSubject(keyword string, option *SearchObjectOption) (*SearchSubjectResult, error) {
	request := c.client.R()
	if option != nil {
		if len(option.ResponseGroup) > 0 {
			request.SetQueryParam("responseGroup", option.ResponseGroup)
		}
		if len(option.Type) > 0 {
			request.SetQueryParam("type", option.Type)
		}
	}
	result := &SearchSubjectResult{}
	response, err := request.SetResult(result).Get("https://api.bgm.tv/search/subject/" + keyword)
	if err != nil {
		return nil, err
	}
	return response.Result().(*SearchSubjectResult), nil
}

var bangumiInfoSource *BangumiInfoSource

type BangumiInfoSource struct {
	client   *BangumiClient
	logScope *youlog.Scope
}

func (s *BangumiInfoSource) SearchMovie(query string) (*SearchMovieResult, error) {
	result, err := s.client.SearchSubject(query, &SearchObjectOption{
		Type:          "2",
		ResponseGroup: "large",
	})
	if err != nil {
		return nil, err
	}
	if result.Results > 0 {
		return &SearchMovieResult{
			Name:    result.List[0].Name,
			Summary: result.List[0].Summary,
			Cover:   result.List[0].Images.Large,
		}, nil
	}
	return nil, nil
}
func (s *BangumiInfoSource) SearchMovieList(query string) ([]*SearchMovieResult, error) {
	result, err := s.client.SearchSubject(query, &SearchObjectOption{
		Type:          "2",
		ResponseGroup: "large",
	})
	if err != nil {
		return nil, err
	}
	resultList := make([]*SearchMovieResult, 0)
	for _, item := range result.List {
		resultList = append(resultList, &SearchMovieResult{
			Name:    item.Name,
			Cover:   item.Images.Large,
			Summary: item.Summary,
		})
	}
	return resultList, nil
}
func (s *BangumiInfoSource) SearchTv(query string) (*SearchTVResult, error) {
	result, err := s.client.SearchSubject(query, &SearchObjectOption{
		Type:          "2",
		ResponseGroup: "large",
	})
	if err != nil {
		return nil, err
	}
	if result.Results > 0 {
		return &SearchTVResult{
			Name:    result.List[0].Name,
			Summary: result.List[0].Summary,
			Cover:   result.List[0].Images.Large,
		}, nil
	}
	return nil, nil
}
func (s *BangumiInfoSource) SearchTvList(query string) ([]*SearchTVResult, error) {
	result, err := s.client.SearchSubject(query, &SearchObjectOption{
		Type:          "2",
		ResponseGroup: "large",
	})
	if err != nil {
		return nil, err
	}
	resultList := make([]*SearchTVResult, 0)
	for _, item := range result.List {
		resultList = append(resultList, &SearchTVResult{
			Name:    item.Name,
			Cover:   item.Images.Large,
			Summary: item.Summary,
		})
	}
	return resultList, nil
}
func (s *BangumiInfoSource) MatchEntity(entity *database.Entity) error {
	result, err := s.SearchMovie(entity.Name)
	if err != nil {
		return err
	}
	if result != nil {
		entity.Cover = result.Cover
		entity.Summary = result.Summary
		return nil
	}
	return nil
}

func (s *BangumiInfoSource) DownloadCover(url string) (string, error) {
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

func InitBangumiInfoSource() {
	logScope := plugin.DefaultYouLogPlugin.Logger.NewScope("bangumi")
	bangumiConfig := config.Instance.BangumiConfig
	logScope.Info("init bangumi info source")
	if !bangumiConfig.Enable {
		logScope.Info("bangumi info source is disabled")
		return
	}
	bangumiInfoSource = &BangumiInfoSource{
		client:   NewBangumiClient(),
		logScope: logScope,
	}
}

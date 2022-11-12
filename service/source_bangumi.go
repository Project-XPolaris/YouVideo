package service

import (
	"context"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
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
type GetSubjectResult struct {
	Date     string `json:"date"`
	Platform string `json:"platform"`
	Images   struct {
		Small  string `json:"small"`
		Grid   string `json:"grid"`
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Common string `json:"common"`
	} `json:"images"`
	Summary string `json:"summary"`
	Name    string `json:"name"`
	NameCn  string `json:"name_cn"`
	Tags    []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tags"`
	Infobox []struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	} `json:"infobox"`
	Rating struct {
		Rank  int `json:"rank"`
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
	} `json:"rating"`
	TotalEpisodes int `json:"total_episodes"`
	Collection    struct {
		OnHold  int `json:"on_hold"`
		Dropped int `json:"dropped"`
		Wish    int `json:"wish"`
		Collect int `json:"collect"`
		Doing   int `json:"doing"`
	} `json:"collection"`
	ID      int  `json:"id"`
	Eps     int  `json:"eps"`
	Volumes int  `json:"volumes"`
	Locked  bool `json:"locked"`
	Nsfw    bool `json:"nsfw"`
	Type    int  `json:"type"`
}

func NewBangumiClient() *BangumiClient {
	client := resty.New()
	client.SetHeader("User-Agent", "allentom/YouVideo (https://github.com/Project-XPolaris/YouVideo)")
	return &BangumiClient{
		client: client,
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
func (c *BangumiClient) GetSubjectById(id string) (*GetSubjectResult, error) {
	request := c.client.R()
	result := &GetSubjectResult{}
	response, err := request.SetResult(result).Get("https://api.bgm.tv/v0/subjects/" + id)
	if err != nil {
		return nil, err
	}
	return response.Result().(*GetSubjectResult), nil
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
func (s *BangumiInfoSource) ApplyEntityById(entity *database.Entity, id string) error {
	s.logScope.Info("apply entity by id ", id)
	result, err := s.client.GetSubjectById(id)
	if err != nil {
		return err
	}
	entity.Name = result.Name
	entity.Summary = result.Summary
	if len(result.Images.Large) > 0 {
		entity.Cover, err = s.DownloadCover(result.Images.Large)
		coverFilename, err := s.DownloadCover(result.Images.Large)
		if err != nil {
			return err
		}
		s.logScope.Info("download cover ", coverFilename)
		entity.Cover = coverFilename
		storage := plugin.GetDefaultStorage()
		reader, err := storage.Get(context.Background(), plugin.GetDefaultBucket(), "entity/"+coverFilename)
		if err != nil {
			return err
		}
		s.logScope.Info("upload cover to storage ", coverFilename)
		width, height, err := util.GetImageSize(reader)
		if err != nil {
			return err
		}
		entity.CoverWidth = width
		entity.CoverHeight = height
		s.logScope.Info("download cover ", entity.Cover, "width", width, "height", height)
	}
	tags := make([]database.EntityTag, 0)
	for _, tag := range result.Tags {
		entityTag := database.EntityTag{
			Name:  "Tag",
			Value: tag.Name,
		}
		err = database.Instance.Where(entityTag).FirstOrCreate(&entityTag).Error
		if err != nil {
			return err
		}
		tags = append(tags, entityTag)
	}
	for _, infobox := range result.Infobox {
		value, isString := infobox.Value.(string)
		if !isString {
			continue
		}
		entityTag := database.EntityTag{
			Name:  infobox.Key,
			Value: value,
		}
		err = database.Instance.Where(entityTag).FirstOrCreate(&entityTag).Error
		if err != nil {
			return err
		}
		tags = append(tags, entityTag)
	}
	return database.Instance.Model(entity).Association("Tags").Append(tags)
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

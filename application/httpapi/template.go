package httpapi

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/allentom/harukap/module/task"
	"github.com/allentom/transcoder/ffmpeg"
	"github.com/project-xpolaris/youplustoolkit/youlibrary"
	"github.com/project-xpolaris/youplustoolkit/youplus"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/youtrans"
	"os"
	"path/filepath"
	"time"
)

const formatTime = "2006-01-02 15:04:05"
const formatDate = "2006-01-02"

type BaseListContainer struct {
	Count    int64       `json:"count"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Result   interface{} `json:"result"`
}

func (t *BaseListContainer) SerializeList(result interface{}, context map[string]interface{}) {
	t.Count = context["count"].(int64)
	t.Page = context["page"].(int)
	t.PageSize = context["pageSize"].(int)
	t.Result = result
}

type BaseLibraryTemplate struct {
	Id      uint   `json:"id"`
	Path    string `json:"path"`
	Name    string `json:"name"`
	DirName string `json:"dir_name"`
}

func (t *BaseLibraryTemplate) Assign(library *database.Library) {
	t.Id = library.ID
	t.Path = library.Path
	t.DirName = filepath.Base(library.Path)
	t.Name = library.Name
}

type BaseSubtitleTemplate struct {
	Id    uint   `json:"id"`
	Label string `json:"label"`
}
type BaseFileTemplate struct {
	Id             uint                   `json:"id"`
	Path           string                 `json:"path"`
	Cover          string                 `json:"cover,omitempty"`
	Duration       float64                `json:"duration"`
	Size           int64                  `json:"size"`
	Bitrate        int64                  `json:"bitrate"`
	MainVideoCodec string                 `json:"main_video_codec，omitempty"`
	MainAudioCodec string                 `json:"main_audio_codec,omitempty"`
	VideoId        uint                   `json:"video_id"`
	Name           string                 `json:"name"`
	CoverWidth     uint                   `json:"coverWidth"`
	CoverHeight    uint                   `json:"coverHeight"`
	Subtitles      []BaseSubtitleTemplate `json:"subtitles,omitempty"`
}

func (t *BaseFileTemplate) Assign(file *database.File) {
	t.Id = file.ID
	t.Path = file.Path
	if len(file.Cover) > 0 {
		t.Cover = fmt.Sprintf("/video/file/%d/cover", file.ID)
	}
	t.VideoId = file.VideoId
	t.Duration = file.Duration
	t.Size = file.Size
	t.Bitrate = file.Bitrate
	t.MainVideoCodec = file.MainVideoCodec
	t.MainAudioCodec = file.MainAudioCodec
	t.Name = filepath.Base(file.Path)
	t.CoverWidth = uint(file.CoverWidth)
	t.CoverHeight = uint(file.CoverHeight)
	t.Subtitles = make([]BaseSubtitleTemplate, 0)
	if file.Subtitles != nil {
		for _, subtitle := range file.Subtitles {
			t.Subtitles = append(t.Subtitles, BaseSubtitleTemplate{
				Id:    subtitle.ID,
				Label: subtitle.Label,
			})
		}
	}
}

type BaseVideoTemplate struct {
	Id        uint                    `json:"id"`
	BaseDir   string                  `json:"base_dir"`
	DirName   string                  `json:"dirName"`
	Name      string                  `json:"name"`
	LibraryId uint                    `json:"library_id"`
	Type      string                  `json:"type"`
	Files     []BaseFileTemplate      `json:"files,omitempty"`
	Infos     []BaseVideoMetaTemplate `json:"infos,omitempty"`
	Subject   *youlibrary.Subject     `json:"subject,omitempty"`
	Release   string                  `json:"release,omitempty"`
	EntityId  *uint                   `json:"entityId,omitempty"`
}

func (t *BaseVideoTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	video := dataModel.(*database.Video)
	t.Assign(video)
	t.DirName = filepath.Base(video.BaseDir)
	return nil
}

func (t *BaseVideoTemplate) Assign(video *database.Video) {
	t.Id = video.ID
	t.BaseDir = video.BaseDir
	t.Name = video.Name
	t.LibraryId = video.LibraryId
	t.Type = video.Type
	t.EntityId = video.EntityID
	if video.Release != nil {
		t.Release = video.Release.Format(formatDate)
	}
	if video.Files != nil {
		fileTemplates := make([]BaseFileTemplate, 0)
		for _, file := range video.Files {
			template := BaseFileTemplate{}
			template.Assign(&file)
			fileTemplates = append(fileTemplates, template)
		}
		t.Files = fileTemplates
	}
	if video.Infos != nil {
		infoTemplates := make([]BaseVideoMetaTemplate, 0)
		for _, info := range video.Infos {
			template := BaseVideoMetaTemplate{}
			template.Serializer(info, nil)
			infoTemplates = append(infoTemplates, template)
		}
		t.Infos = infoTemplates
	}
}

type BaseFileItemTemplate struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (t *BaseFileItemTemplate) Assign(info os.FileInfo, rootPath string) {
	if info.IsDir() {
		t.Type = "Directory"
	} else {
		t.Type = "File"
	}
	t.Name = info.Name()
	t.Path = filepath.Join(rootPath, info.Name())
}
func (t *BaseFileItemTemplate) AssignWithYouPlusItem(item youplus.ReadDirItem) {
	t.Type = item.Type
	t.Path = item.Path
	t.Name = filepath.Base(item.Path)
}

type BaseTaskTemplate struct {
	Id     string      `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Output interface{} `json:"output"`
}

func (t *BaseTaskTemplate) Assign(taskData task.Task) {
	t.Id = taskData.GetId()
	t.Status = taskData.GetStatus()
	t.Type = taskData.GetType()
	output, _ := taskData.Output()
	t.Output = output
}

func (t *BaseTaskTemplate) AssignWithTrans(task youtrans.TaskResponse) {
	t.Id = task.Id
	t.Status = task.Status
	t.Type = "transcode"
	t.Output = haruka.JSON{
		"input":   task.Input,
		"output":  task.Output,
		"process": task.Process,
	}
}

type BaseTagTemplate struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

func (t *BaseTagTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	tagModel := dataModel.(*database.Tag)
	t.Id = tagModel.ID
	t.Name = tagModel.Name
	return nil
}

type BaseCodecTemplate struct {
	Name string   `json:"name"`
	Desc string   `json:"desc"`
	Type string   `json:"type"`
	Feat []string `json:"feat"`
}

func (t *BaseCodecTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(ffmpeg.Codec)
	t.Name = model.Name
	t.Desc = model.Desc
	if model.Flags.AudioCodec {
		t.Type = "Audio"
	}
	if model.Flags.VideoCodec {
		t.Type = "Video"
	}
	if model.Flags.SubtitleCodec {
		t.Type = "Subtitle"
	}
	t.Feat = []string{}
	if model.Flags.Decoding {
		t.Feat = append(t.Feat, "decode")
	}
	if model.Flags.Encoding {
		t.Feat = append(t.Feat, "encode")
	}
	return nil
}

type BaseFormatTemplate struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func (t *BaseFormatTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(ffmpeg.SupportFormat)
	t.Name = model.Name
	t.Desc = model.Desc
	return nil
}

type BaseHistoryTemplate struct {
	VideoId uint   `json:"video_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Cover   string `json:"cover,omitempty"`
	Time    string `json:"time,omitempty"`
}

func (t *BaseHistoryTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(*database.History)
	if model.Video != nil {
		t.Name = model.Video.Name
		if model.Video.Files != nil && len(model.Video.Files) > 0 {
			t.Cover = fmt.Sprintf("/video/file/%d/cover", model.Video.Files[0].ID)
		}
	}
	t.Time = model.UpdatedAt.Format(formatTime)
	t.VideoId = model.Video.ID
	return nil
}

type BaseFolderTemplate struct {
	Id     uint        `json:"id"`
	Name   string      `json:"name"`
	Videos interface{} `json:"videos"`
}

func (t *BaseFolderTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(*database.Folder)
	t.Id = model.ID
	t.Name = filepath.Base(model.Path)
	data := serializer.SerializeMultipleTemplate(model.Videos, &BaseVideoTemplate{}, map[string]interface{}{})
	t.Videos = data
	return nil
}

type BaseVideoMetaTemplate struct {
	Id    uint   `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (t *BaseVideoMetaTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(*database.VideoMetaItem)
	t.Id = model.ID
	t.Key = model.Key
	t.Value = model.Value
	return nil
}

type BaseEntityTemplate struct {
	Id          uint                    `json:"id"`
	Name        string                  `json:"name"`
	Summary     string                  `json:"summary"`
	Videos      []BaseVideoTemplate     `json:"videos,omitempty"`
	Cover       string                  `json:"cover,omitempty"`
	CoverWidth  int                     `json:"coverWidth,omitempty"`
	CoverHeight int                     `json:"coverHeight,omitempty"`
	Infos       []BaseVideoMetaTemplate `json:"infos,omitempty"`
	Release     string                  `json:"release,omitempty"`
	LibraryId   uint                    `json:"libraryId,omitempty"`
}

func (t *BaseEntityTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(*database.Entity)
	t.Id = model.ID
	t.Name = model.Name
	t.LibraryId = model.LibraryId
	t.Summary = model.Summary
	var release *time.Time
	if model.Videos != nil {
		videoTemplates := make([]BaseVideoTemplate, 0)
		cover := ""
		coverSelector := "auto"
		coverWidth := int64(model.CoverWidth)
		coverHeight := int64(model.CoverHeight)
		hasCover := false
		if len(model.Cover) > 0 {
			cover = fmt.Sprintf("/entity/%d/cover", model.ID)
			hasCover = true
		}
		infos := make([]*database.VideoMetaItem, 0)
		for _, video := range model.Videos {
			videoTemplate := BaseVideoTemplate{}
			videoTemplate.Serializer(video, map[string]interface{}{})
			videoTemplates = append(videoTemplates, videoTemplate)
			if video.Files != nil && coverSelector == "auto" && !hasCover {
				for _, file := range video.Files {
					if len(file.Cover) > 0 {
						cover = fmt.Sprintf("/video/file/%d/cover", file.ID)
						coverWidth = file.CoverWidth
						coverHeight = file.CoverHeight
						if file.AutoGenCover {
							coverSelector = "auto"
						} else {
							coverSelector = "cover"
						}
					}
				}
			}

			if video.Infos != nil {
				if video.Release != nil {
					if release == nil {
						release = video.Release
					} else {
						if video.Release.Before(*release) {
							release = video.Release
						}
					}
				}
				for _, info := range video.Infos {
					linq.From(video.Infos).WhereT(func(item *database.VideoMetaItem) bool {
						return item.ID != info.ID
					}).ToSlice(&infos)
					infos = append(infos, info)
				}
			}
		}

		if len(infos) > 0 {
			infosTemplate := make([]BaseVideoMetaTemplate, 0)
			for _, info := range infos {
				infoTemplate := BaseVideoMetaTemplate{}
				infoTemplate.Serializer(info, map[string]interface{}{})
				infosTemplate = append(infosTemplate, infoTemplate)
			}
			t.Infos = infosTemplate
		}
		t.Videos = videoTemplates
		t.Cover = cover
		t.CoverWidth = int(coverWidth)
		t.CoverHeight = int(coverHeight)
		// get cover
	}
	if release != nil {
		t.Release = release.Format(formatDate)
	}
	return nil
}

type CCTemplate struct {
	Index int    `json:"index"`
	Start int64  `json:"start"`
	End   int64  `json:"end"`
	Text  string `json:"text"`
}

func NewCCTemplate(cc *service.CC) *CCTemplate {
	return &CCTemplate{
		Index: cc.Index,
		Start: cc.StartTime.Milliseconds(),
		End:   cc.EndTime.Milliseconds(),
		Text:  cc.Text,
	}
}
func NewCCTemplates(ccs []*service.CC) []*CCTemplate {
	templates := make([]*CCTemplate, 0)
	for _, cc := range ccs {
		templates = append(templates, NewCCTemplate(cc))
	}
	return templates
}

type SearchMoveInformationTemplate struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Cover   string `json:"cover"`
	Source  string `json:"source"`
}

func NewSearchMoveInformationTemplate(m *service.SearchMovieResult, source string) *SearchMoveInformationTemplate {
	data := &SearchMoveInformationTemplate{
		Name:    m.Name,
		Summary: m.Summary,
		Cover:   m.Cover,
	}
	switch source {
	case "tmdb":
		data.Cover = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", m.Cover)

	}
	return data
}

func NewSearchMoveInformationTemplates(ms []*service.SearchMovieResult, source string) []*SearchMoveInformationTemplate {
	templates := make([]*SearchMoveInformationTemplate, 0)
	for _, m := range ms {
		templates = append(templates, NewSearchMoveInformationTemplate(m, source))
	}
	return templates
}

type SearchTvInformationTemplate struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Cover   string `json:"cover"`
	Source  string `json:"source"`
}

func NewSearchTvInformationTemplate(m *service.SearchTVResult, source string) *SearchTvInformationTemplate {
	data := &SearchTvInformationTemplate{
		Name:    m.Name,
		Summary: m.Summary,
		Cover:   m.Cover,
		Source:  source,
	}
	switch source {
	case "tmdb":
		data.Cover = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", m.Cover)

	}
	return data
}

func NewSearchTvInformationTemplates(ms []*service.SearchTVResult, source string) []*SearchTvInformationTemplate {
	templates := make([]*SearchTvInformationTemplate, 0)
	for _, m := range ms {
		templates = append(templates, NewSearchTvInformationTemplate(m, source))
	}
	return templates
}

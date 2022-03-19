package httpapi

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/allentom/transcoder/ffmpeg"
	"github.com/project-xpolaris/youplustoolkit/youlibrary"
	"github.com/project-xpolaris/youplustoolkit/youplus"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service/task"
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

type BaseFileTemplate struct {
	Id             uint    `json:"id"`
	Path           string  `json:"path"`
	Cover          string  `json:"cover,omitempty"`
	Duration       float64 `json:"duration"`
	Size           int64   `json:"size"`
	Bitrate        int64   `json:"bitrate"`
	MainVideoCodec string  `json:"main_video_codec，omitempty"`
	MainAudioCodec string  `json:"main_audio_codec,omitempty"`
	VideoId        uint    `json:"video_id"`
	Name           string  `json:"name"`
	CoverWidth     uint    `json:"coverWidth"`
	CoverHeight    uint    `json:"coverHeight"`
	Subtitles      string  `json:"subtitles,omitempty"`
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
	t.Subtitles = file.Subtitles
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
	EntityId  uint                    `json:"entityId,omitempty"`
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

func (t *BaseTaskTemplate) Assign(taskData *task.Task) {
	t.Id = taskData.Id
	t.Status = task.TaskStatusNameMapping[taskData.Status]
	t.Type = task.TaskTypeNameMapping[taskData.Type]
	t.Output = taskData.Output
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
	Videos      []BaseVideoTemplate     `json:"videos,omitempty"`
	Cover       string                  `json:"cover,omitempty"`
	CoverWidth  int                     `json:"coverWidth,omitempty"`
	CoverHeight int                     `json:"coverHeight,omitempty"`
	Infos       []BaseVideoMetaTemplate `json:"infos,omitempty"`
	Release     string                  `json:"release,omitempty"`
}

func (t *BaseEntityTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(*database.Entity)
	t.Id = model.ID
	t.Name = model.Name
	var release *time.Time
	if model.Videos != nil {
		videoTemplates := make([]BaseVideoTemplate, 0)
		cover := ""
		coverSelector := "auto"
		var coverWidth int64
		var coverHeight int64
		infos := make([]*database.VideoMetaItem, 0)
		for _, video := range model.Videos {
			videoTemplate := BaseVideoTemplate{}
			videoTemplate.Serializer(video, map[string]interface{}{})
			videoTemplates = append(videoTemplates, videoTemplate)
			if video.Files != nil && coverSelector == "auto" {
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
package application

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/allentom/transcoder/ffmpeg"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/youplus"
	"github.com/projectxpolaris/youvideo/youtrans"
	"os"
	"path/filepath"
)

const formatTime = "2006-01-02 15:04:05"

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
	Subtitles      string  `json:"subtitles,omitempty"`
}

func (t *BaseFileTemplate) Assign(file *database.File) {
	t.Id = file.ID
	t.Path = file.Path
	if len(file.Cover) > 0 {
		t.Cover = fmt.Sprintf("/covers/%s", file.Cover)
	}
	t.VideoId = file.VideoId
	t.Duration = file.Duration
	t.Size = file.Size
	t.Bitrate = file.Bitrate
	t.MainVideoCodec = file.MainVideoCodec
	t.MainAudioCodec = file.MainAudioCodec
	t.Name = filepath.Base(file.Path)
	t.Subtitles = file.Subtitles
}

type BaseVideoTemplate struct {
	Id        uint               `json:"id"`
	BaseDir   string             `json:"base_dir"`
	DirName   string             `json:"dirName"`
	Name      string             `json:"name"`
	LibraryId uint               `json:"library_id"`
	Files     []BaseFileTemplate `json:"files,omitempty"`
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
	if video.Files != nil {
		fileTemplates := make([]BaseFileTemplate, 0)
		for _, file := range video.Files {
			template := BaseFileTemplate{}
			template.Assign(&file)
			fileTemplates = append(fileTemplates, template)
		}
		t.Files = fileTemplates
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

func (t *BaseTaskTemplate) Assign(task *service.Task) {
	t.Id = task.Id
	t.Status = service.TaskStatusNameMapping[task.Status]
	t.Type = service.TaskTypeNameMapping[task.Type]
	t.Output = task.Output
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
			t.Cover = fmt.Sprintf("/covers/%s", model.Video.Files[0].Cover)
		}
	}
	t.Time = model.UpdatedAt.Format(formatTime)
	t.VideoId = model.Video.ID
	return nil
}

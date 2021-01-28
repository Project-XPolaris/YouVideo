package application

import (
	"fmt"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"os"
	"path/filepath"
)

type BaseLibraryTemplate struct {
	Id      uint   `json:"id"`
	Path    string `json:"path"`
	DirName string `json:"dir_name"`
}

func (t *BaseLibraryTemplate) Assign(library *database.Library) {
	t.Id = library.ID
	t.Path = library.Path
	t.DirName = filepath.Base(library.Path)
}

type BaseFileTemplate struct {
	Id             uint    `json:"id"`
	Path           string  `json:"path"`
	Cover          string  `json:"cover,omitempty"`
	Duration       float64 `json:"duration"`
	Size           int64   `json:"size"`
	Bitrate        int64   `json:"bitrate"`
	MainVideoCodec string  `json:"main_video_codec"`
	MainAudioCodec string  `json:"main_audio_codec"`
	VideoId        uint    `json:"video_id"`
	Name           string  `json:"name"`
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
}

type BaseVideoTemplate struct {
	Id        uint               `json:"id"`
	BaseDir   string             `json:"base_dir"`
	Name      string             `json:"name"`
	LibraryId uint               `json:"library_id"`
	Files     []BaseFileTemplate `json:"files,omitempty"`
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

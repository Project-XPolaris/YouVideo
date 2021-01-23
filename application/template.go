package application

import (
	"fmt"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"os"
	"path/filepath"
	"strings"
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

type BaseVideoTemplate struct {
	Id        uint   `json:"id"`
	Path      string `json:"path"`
	Cover     string `json:"cover,omitempty"`
	LibraryId uint   `json:"library_id"`
	Name      string `json:"name"`
}

func (t *BaseVideoTemplate) Assign(video *database.Video) {
	t.Id = video.ID
	t.Path = video.Path
	if len(video.Cover) > 0 {
		t.Cover = fmt.Sprintf("/covers/%s", video.Cover)
	}
	baseFileName := filepath.Base(video.Path)
	extension := filepath.Ext(baseFileName)
	t.Name = strings.Replace(baseFileName, extension, "", 1)
	t.LibraryId = video.LibraryId
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

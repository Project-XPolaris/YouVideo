package application

import (
	"fmt"
	"github.com/projectxpolaris/youvideo/database"
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

type BaseVideoTemplate struct {
	Id        uint   `json:"id"`
	Path      string `json:"path"`
	Cover     string `json:"cover,omitempty"`
	LibraryId uint   `json:"library_id"`
}

func (t *BaseVideoTemplate) Assign(video *database.Video) {
	t.Id = video.ID
	t.Path = video.Path
	if len(video.Cover) > 0 {
		t.Cover = fmt.Sprintf("/covers/%s", video.Cover)
	}
	t.LibraryId = video.LibraryId
}

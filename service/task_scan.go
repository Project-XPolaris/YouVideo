package service

import (
	"github.com/projectxpolaris/youvideo/database"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type ScanTaskOutput struct {
	Id          uint   `json:"id"`
	Path        string `json:"path"`
	Total       int64  `json:"total"`
	Current     int64  `json:"current"`
	CurrentPath string `json:"currentPath"`
	CurrentName string `json:"currentName"`
}

func CreateSyncLibraryTask(library *database.Library) *Task {
	for _, task := range DefaultTaskPool.Tasks {
		if task.Output.(*ScanTaskOutput).Id == library.ID {
			if task.Status == TaskStatusRunning {
				return task
			}
			DefaultTaskPool.RemoveTaskById(task.Id)
			break
		}
	}
	id := xid.New().String()
	output := &ScanTaskOutput{
		Id:   library.ID,
		Path: library.Path,
	}
	task := &Task{
		Id:     id,
		Type:   TaskTypeScanLibrary,
		Status: TaskStatusRunning,
		Output: output,
	}
	logger := TaskLogger.WithFields(logrus.Fields{
		"id":        id,
		"path":      library.Path,
		"libraryId": library.ID,
	})
	go func() {
		logger.Info("task start")
		err := CheckLibrary(library.ID)
		if err != nil {
			task.SetError(err)
			return
		}
		pathList, err := ScanVideo(library)
		if err != nil {
			task.SetError(err)
			return
		}
		output.Total = int64(len(pathList))
		for idx, path := range pathList {
			output.Current = int64(idx + 1)
			output.CurrentPath = path
			output.CurrentName = filepath.Base(path)
			err = CreateVideoFile(path, library.ID)
			if err != nil {
				logger.Error(err)
			}
		}
		task.Status = TaskStatusDone
	}()
	DefaultTaskPool.Lock()
	DefaultTaskPool.Tasks = append(DefaultTaskPool.Tasks, task)
	DefaultTaskPool.Unlock()
	return task
}

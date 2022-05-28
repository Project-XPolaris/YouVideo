package task

import (
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
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
type CreateScanTaskOption struct {
	LibraryId      uint
	Uid            string
	MatchSubject   bool
	OnFileComplete func(task *Task)
	OnFileError    func(task *Task, err error)
	OnError        func(task *Task, err error)
	OnComplete     func(task *Task)
}

func CreateSyncLibraryTask(option CreateScanTaskOption) (*Task, error) {
	for _, task := range DefaultTaskPool.Tasks {
		if scanOutput, ok := task.Output.(*ScanTaskOutput); ok && scanOutput.Id == option.LibraryId {
			if task.Status == TaskStatusRunning {
				return task, nil
			}
			DefaultTaskPool.RemoveTaskById(task.Id)
			break
		}
	}
	var library database.Library
	err := database.Instance.Preload("Users").Find(&library, option.LibraryId).Error
	if err != nil {
		return nil, err
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
		Uid:    option.Uid,
	}
	logger := TaskLogger.WithFields(logrus.Fields{
		"id":        id,
		"path":      library.Path,
		"libraryId": library.ID,
	})
	go func() {
		logger.Info("task start")
		// remove file where is not found
		err := service.CheckLibrary(library.ID)
		if err != nil {
			task.SetError(err)
			if option.OnError != nil {
				option.OnError(task, err)
			}
			return
		}
		pathList, err := service.ScanVideo(&library)
		if err != nil {
			task.SetError(err)
			if option.OnError != nil {
				option.OnError(task, err)
			}
			return
		}
		output.Total = int64(len(pathList))
		for idx, path := range pathList {
			output.Current = int64(idx + 1)
			output.CurrentPath = path
			output.CurrentName = filepath.Base(path)
			err = service.CreateVideoFile(path, library.ID, library.DefaultVideoType, option.MatchSubject)
			if err != nil {
				if option.OnFileError != nil {
					option.OnFileError(task, err)
				}
				logger.Error(err)
			} else {
				if option.OnFileComplete != nil {
					option.OnFileComplete(task)
				}
			}

		}
		task.Status = TaskStatusDone
		if option.OnComplete != nil {
			option.OnComplete(task)
		}
	}()
	DefaultTaskPool.Lock()
	DefaultTaskPool.Tasks = append(DefaultTaskPool.Tasks, task)
	DefaultTaskPool.Unlock()
	return task, nil
}

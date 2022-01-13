package task

import (
	"errors"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type GenerateVideoMetaTaskOutput struct {
	Id          uint             `json:"id"`
	Total       int64            `json:"total"`
	Current     int64            `json:"current"`
	CurrentPath string           `json:"currentPath"`
	CurrentName string           `json:"currentName"`
	Library     database.Library `json:"-"`
}
type CreateGenerateMetaOption struct {
	LibraryId       uint
	Uid             string
	OnVideoComplete func(task *Task)
	OnFileComplete  func(task *Task)
	OnFileError     func(task *Task, err error)
	OnComplete      func(task *Task)
}

func CreateGenerateVideoMetaTask(option CreateGenerateMetaOption) (*Task, error) {
	for _, task := range DefaultTaskPool.Tasks {
		if task.Output.(*GenerateVideoMetaTaskOutput).Id == option.LibraryId {
			if task.Status == TaskStatusRunning {
				return task, nil
			}
			// recreate task
			DefaultTaskPool.RemoveTaskById(task.Id)
			break
		}
	}
	if !service.DefaultLibraryLockManager.TryToLock(option.LibraryId) {
		return nil, errors.New("library is busy")
	}
	output := &GenerateVideoMetaTaskOutput{
		Id: option.LibraryId,
	}
	task := &Task{
		Id:     xid.New().String(),
		Type:   TaskTypeMeta,
		Status: TaskStatusRunning,
		Output: output,
		Uid:    option.Uid,
	}
	var library database.Library
	err := database.Instance.Where("id = ?", option.LibraryId).Preload("Videos").Preload("Videos.Files").Find(&library).Error
	if err != nil {
		service.DefaultLibraryLockManager.UnlockLibrary(option.LibraryId)
		return nil, err
	}
	output.Library = library
	output.Total = int64(len(library.Videos))

	go func() {
		for idx, video := range library.Videos {
			output.Current = int64(idx) + 1
			output.CurrentName = video.Name
			for _, file := range video.Files {
				doneChan := make(chan struct{}, 0)
				errChan := make(chan error, 0)
				output.CurrentPath = file.Path
				service.DefaultVideoMetaAnalyzer.In <- service.VideoMetaAnalyzerInput{
					File:    &file,
					OnDone:  doneChan,
					OnError: errChan,
				}
				select {
				case <-doneChan:
				case metaErr := <-errChan:
					if option.OnFileError != nil {
						option.OnFileError(task, metaErr)
					}
					logrus.Error(metaErr)
				}
				if option.OnFileComplete != nil {
					option.OnFileComplete(task)
				}
			}
			if option.OnVideoComplete != nil {
				option.OnVideoComplete(task)
			}
		}
		if option.OnComplete != nil {
			option.OnComplete(task)
		}
		task.Status = TaskStatusDone
		service.DefaultLibraryLockManager.UnlockLibrary(library.ID)
	}()
	DefaultTaskPool.Lock()
	DefaultTaskPool.Tasks = append(DefaultTaskPool.Tasks, task)
	DefaultTaskPool.Unlock()
	return task, nil
}

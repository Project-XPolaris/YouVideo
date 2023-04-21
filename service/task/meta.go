package task

import (
	"errors"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/sirupsen/logrus"
)

type GenerateVideoMetaTask struct {
	task.BaseTask
	TaskOutput *GenerateVideoMetaTaskOutput
	Library    database.Library
	Option     *CreateGenerateMetaOption
}

func (t *GenerateVideoMetaTask) Stop() error {
	return nil
}

func (t *GenerateVideoMetaTask) Start() error {
	for idx, video := range t.Library.Videos {
		t.TaskOutput.Current = int64(idx) + 1
		t.TaskOutput.CurrentName = video.Name
		for _, file := range video.Files {
			doneChan := make(chan struct{}, 0)
			errChan := make(chan error, 0)
			t.TaskOutput.CurrentPath = file.Path
			service.DefaultVideoMetaAnalyzer.In <- service.VideoMetaAnalyzerInput{
				File:    &file,
				OnDone:  doneChan,
				OnError: errChan,
			}
			select {
			case <-doneChan:
			case metaErr := <-errChan:
				if t.Option.OnFileError != nil {
					t.Option.OnFileError(t, metaErr)
				}
				logrus.Error(metaErr)
			}
			if t.Option.OnFileComplete != nil {
				t.Option.OnFileComplete(t)
			}
		}
		if t.Option.OnVideoComplete != nil {
			t.Option.OnVideoComplete(t)
		}
	}
	if t.Option.OnComplete != nil {
		t.Option.OnComplete(t)
	}
	t.Done()
	service.DefaultLibraryLockManager.UnlockLibrary(t.Library.ID)
	return nil
}

func (t *GenerateVideoMetaTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

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
	OnVideoComplete func(task *GenerateVideoMetaTask)
	OnFileComplete  func(task *GenerateVideoMetaTask)
	OnFileError     func(task *GenerateVideoMetaTask, err error)
	OnComplete      func(task *GenerateVideoMetaTask)
}

func CreateGenerateVideoMetaTask(option CreateGenerateMetaOption) (*GenerateVideoMetaTask, error) {
	existRunningTask := module.TaskModule.Pool.GetTaskWithStatus(TaskTypeNameMapping[TaskTypeMeta], task.GetStatusText(nil, task.StatusRunning))
	if existRunningTask != nil {
		return existRunningTask.(*GenerateVideoMetaTask), nil
	}
	if !service.DefaultLibraryLockManager.TryToLock(option.LibraryId) {
		return nil, errors.New("library is busy")
	}
	output := &GenerateVideoMetaTaskOutput{
		Id: option.LibraryId,
	}
	task := &GenerateVideoMetaTask{
		BaseTask:   *task.NewBaseTask(TaskTypeNameMapping[TaskTypeMeta], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		TaskOutput: output,
		Option:     &option,
	}
	var library database.Library
	err := database.Instance.Where("id = ?", option.LibraryId).Preload("Videos").Preload("Videos.Files").Find(&library).Error
	if err != nil {
		service.DefaultLibraryLockManager.UnlockLibrary(option.LibraryId)
		return nil, err
	}
	task.Library = library
	output.Library = library
	output.Total = int64(len(library.Videos))
	module.TaskModule.Pool.AddTask(task)
	return task, nil
}

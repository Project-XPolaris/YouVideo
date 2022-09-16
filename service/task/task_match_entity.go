package task

import (
	"errors"
	"fmt"
	"github.com/allentom/harukap/module/task"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
)

type MatchEntityTask struct {
	task.BaseTask
	TaskOutput *MatchEntityTaskOutput
	Library    database.Library
	Option     *MatchEntityOption
	Logger     *youlog.Scope
}

func (t *MatchEntityTask) Stop() error {
	return nil
}

func (t *MatchEntityTask) Start() error {
	for idx, entity := range t.Library.Entity {
		t.TaskOutput.Current = int64(idx) + 1
		t.TaskOutput.CurrentName = entity.Name
		t.Logger.Info(fmt.Sprintf("match entity for [%s]", entity.Name))
		source := service.GetInfoSource(t.Option.Source)
		err := source.MatchEntity(entity)
		if err != nil {
			if t.Option.OnEntityError != nil {
				t.Option.OnEntityError(t, err)
			}
			continue
		}
		if entity.Cover != "" {
			coverFilename, err := source.DownloadCover(entity.Cover)
			if err != nil {
				if t.Option.OnEntityError != nil {
					t.Option.OnEntityError(t, err)
				}
				entity.Cover = ""
			} else {
				entity.Cover = coverFilename
			}
		}
		err = database.Instance.Save(&entity).Error
		if err != nil {
			if t.Option.OnEntityError != nil {
				t.Option.OnEntityError(t, err)
			}
			continue
		}
	}
	if t.Option.OnComplete != nil {
		t.Option.OnComplete(t)
	}
	t.BaseTask.Status = TaskStatusNameMapping[TaskStatusDone]
	service.DefaultLibraryLockManager.UnlockLibrary(t.Library.ID)
	return nil
}

func (t *MatchEntityTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

type MatchEntityTaskOutput struct {
	Id          uint             `json:"id"`
	Total       int64            `json:"total"`
	Current     int64            `json:"current"`
	CurrentName string           `json:"currentName"`
	Library     database.Library `json:"-"`
}
type MatchEntityOption struct {
	LibraryId        uint
	Uid              string
	Source           string
	OnEntityComplete func(task *MatchEntityTask)
	OnEntityError    func(task *MatchEntityTask, err error)
	OnComplete       func(task *MatchEntityTask)
}

func CreateMatchEntityTask(option MatchEntityOption) (*MatchEntityTask, error) {
	existRunningTask := module.TaskModule.Pool.GetTaskWithStatus(TaskTypeNameMapping[TaskTypeMatchEntity], TaskStatusNameMapping[TaskStatusRunning])
	if existRunningTask != nil {
		return existRunningTask.(*MatchEntityTask), nil
	}
	if !service.DefaultLibraryLockManager.TryToLock(option.LibraryId) {
		return nil, errors.New("library is busy")
	}
	output := &MatchEntityTaskOutput{
		Id: option.LibraryId,
	}

	task := &MatchEntityTask{
		BaseTask:   *task.NewBaseTask(TaskTypeNameMapping[TaskTypeMatchEntity], option.Uid, TaskStatusNameMapping[TaskStatusRunning]),
		TaskOutput: output,
		Option:     &option,
	}
	logScope := plugin.DefaultYouLogPlugin.Logger.NewScope("MatchEntityTask").WithFields(youlog.Fields{
		"id": task.Id,
	})
	task.Logger = logScope
	var library database.Library
	err := database.Instance.Where("id = ?", option.LibraryId).Preload("Entity").Find(&library).Error
	if err != nil {
		service.DefaultLibraryLockManager.UnlockLibrary(option.LibraryId)
		return nil, err
	}
	task.Library = library
	output.Library = library
	output.Total = int64(len(library.Entity))
	module.TaskModule.Pool.AddTask(task)
	return task, nil
}

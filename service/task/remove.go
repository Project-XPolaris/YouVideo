package task

import (
	"errors"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
	"github.com/projectxpolaris/youvideo/service"
	"gorm.io/gorm"
)

type RemoveLibraryTask struct {
	task.BaseTask
	TaskOutput *RemoveLibraryOutput
	Library    database.Library
	Option     *RemoveLibraryOption
}

func (t *RemoveLibraryTask) Stop() error {
	return nil
}

func (t *RemoveLibraryTask) Start() error {
	var videos []database.Video
	err := database.Instance.
		Model(&database.Library{Model: gorm.Model{ID: t.Library.ID}}).
		Association("Videos").
		Find(&videos)
	if err != nil {
		if t.Option.OnError != nil {
			t.Option.OnError(t, err)
		}
		return nil
	}
	for _, video := range videos {
		err = service.DeleteVideoById(video.ID)
		if err != nil {
			if t.Option.OnError != nil {
				t.Option.OnError(t, err)
			}
			return nil
		}
	}
	// clear library users
	err = database.Instance.
		Model(&database.Library{Model: gorm.Model{ID: t.Library.ID}}).
		Association("Users").Clear()
	if err != nil {
		if t.Option.OnError != nil {
			t.Option.OnError(t, err)
		}
		return nil
	}
	// clear entity
	var entities []database.Entity
	err = database.Instance.
		Model(&database.Library{Model: gorm.Model{ID: t.Library.ID}}).
		Association("Entity").
		Find(&entities)
	if err != nil {
		if t.Option.OnError != nil {
			t.Option.OnError(t, err)
		}
		return nil
	}
	for _, entity := range entities {
		err = database.Instance.Model(&entity).Association("Tags").Clear()
		if err != nil {
			if t.Option.OnError != nil {
				t.Option.OnError(t, err)
			}
			return nil
		}
	}
	// clear library folder
	err = database.Instance.Unscoped().Where("library_id = ?", t.Library.ID).Delete(database.Folder{}).Error
	if err != nil {
		if t.Option.OnError != nil {
			t.Option.OnError(t, err)
		}
		return nil
	}
	err = database.Instance.Unscoped().Delete(&database.Library{}, t.Library.ID).Error
	if err != nil {
		if t.Option.OnError != nil {
			t.Option.OnError(t, err)
		}
		return nil
	}
	// remove index
	if service.DefaultMeilisearchEngine.Enable {
		err = service.DefaultMeilisearchEngine.DeleteAllIndexByLibrary(t.Library.ID)
		if err != nil {
			if t.Option.OnError != nil {
				t.Option.OnError(t, err)
			}
			return nil
		}
	}
	t.Done()
	if t.Option.OnComplete != nil {
		t.Option.OnComplete(t)
	}
	return nil
}

func (t *RemoveLibraryTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

type RemoveLibraryOutput struct {
	Id   uint   `json:"id"`
	Path string `json:"path"`
}
type RemoveLibraryOption struct {
	LibraryId  uint
	Uid        string
	OnError    func(task *RemoveLibraryTask, err error)
	OnComplete func(task *RemoveLibraryTask)
}

func CreateRemoveLibraryTask(option RemoveLibraryOption) (*RemoveLibraryTask, error) {
	existRunningTask := module.TaskModule.Pool.GetTaskWithStatus(TaskTypeNameMapping[TaskTypeRemove], task.GetStatusText(nil, task.StatusRunning))
	if existRunningTask != nil {
		return existRunningTask.(*RemoveLibraryTask), nil
	}
	if !service.DefaultLibraryLockManager.TryToLock(option.LibraryId) {
		return nil, errors.New("library is busy")
	}
	var library database.Library
	err := database.Instance.Preload("Users").Find(&library, option.LibraryId).Error
	if err != nil {
		return nil, err
	}
	output := &RemoveLibraryOutput{
		Id:   library.ID,
		Path: library.Path,
	}
	task := &RemoveLibraryTask{
		BaseTask:   *task.NewBaseTask(TaskTypeNameMapping[TaskTypeRemove], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		TaskOutput: output,
		Library:    library,
		Option:     &option,
	}
	module.TaskModule.Pool.AddTask(task)
	return task, nil
}

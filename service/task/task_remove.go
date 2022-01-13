package task

import (
	"errors"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

type RemoveLibraryOutput struct {
	Id   uint   `json:"id"`
	Path string `json:"path"`
}
type RemoveLibraryOption struct {
	LibraryId  uint
	Uid        string
	OnError    func(task *Task, err error)
	OnComplete func(task *Task)
}

func CreateRemoveLibraryTask(option RemoveLibraryOption) (*Task, error) {
	for _, task := range DefaultTaskPool.Tasks {
		if removeTask, ok := task.Output.(*RemoveLibraryOutput); ok && removeTask.Id == option.LibraryId {
			if task.Status == TaskStatusRunning {
				return task, nil
			}
			DefaultTaskPool.RemoveTaskById(task.Id)
			break
		}
	}
	if !service.DefaultLibraryLockManager.TryToLock(option.LibraryId) {
		return nil, errors.New("library is busy")
	}
	var library database.Library
	err := database.Instance.Preload("Users").Find(&library, option.LibraryId).Error
	if err != nil {
		return nil, err
	}
	id := xid.New().String()
	output := &RemoveLibraryOutput{
		Id:   library.ID,
		Path: library.Path,
	}
	task := &Task{
		Id:     id,
		Type:   TaskTypeRemove,
		Status: TaskStatusRunning,
		Output: output,
		Uid:    option.Uid,
	}
	go func() {
		var videos []database.Video
		err = database.Instance.
			Model(&database.Library{Model: gorm.Model{ID: library.ID}}).
			Association("Videos").
			Find(&videos)
		if err != nil {
			if option.OnError != nil {
				option.OnError(task, err)
			}
			return
		}
		for _, video := range videos {
			err = service.DeleteVideoById(video.ID)
			if err != nil {
				if option.OnError != nil {
					option.OnError(task, err)
				}
				return
			}
		}
		err = database.Instance.
			Model(&database.Library{Model: gorm.Model{ID: library.ID}}).
			Association("Users").Clear()
		if err != nil {
			if option.OnError != nil {
				option.OnError(task, err)
			}
			return
		}
		err = database.Instance.Unscoped().Delete(&database.Library{}, library.ID).Error
		if err != nil {
			if option.OnError != nil {
				option.OnError(task, err)
			}
			return
		}
		err = database.Instance.Unscoped().Where("library_id = ?", library.ID).Delete(database.Folder{}).Error
		if err != nil {
			if option.OnError != nil {
				option.OnError(task, err)
			}
			return
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

package task

import (
	"errors"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/sirupsen/logrus"
)

type SyncIndexTask struct {
	task.BaseTask
	TaskOutput *SyncTaskOutput
	Library    *database.Library
	Option     *CreateSyncIndexTaskOption
	logger     *logrus.Entry
}
type CreateSyncIndexTaskOption struct {
	LibraryId uint
	Uid       string
}
type SyncTaskOutput struct {
}

func (t *SyncIndexTask) Stop() error {
	return nil
}

func (t *SyncIndexTask) Start() error {
	return service.DefaultMeilisearchEngine.Sync(t.Library.ID)
}

func (t *SyncIndexTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}
func CreateSyncIndexTask(option CreateSyncIndexTaskOption) (*SyncIndexTask, error) {
	if service.DefaultMeilisearchEngine.Enable {
		return nil, errors.New("meilisearch engine is not enable")
	}
	existRunningTask := module.TaskModule.Pool.GetTaskWithStatus(TaskTypeNameMapping[TaskTypeSyncIndex], TaskStatusNameMapping[TaskStatusRunning])
	if existRunningTask != nil {
		return existRunningTask.(*SyncIndexTask), nil
	}
	if service.DefaultLibraryLockManager.IsLock(option.LibraryId) {
		return nil, errors.New("library is busy")
	}
	var library database.Library
	err := database.Instance.Find(&library, option.LibraryId).Error
	if err != nil {
		return nil, err
	}
	output := &SyncTaskOutput{}
	task := &SyncIndexTask{
		BaseTask:   *task.NewBaseTask(TaskTypeNameMapping[TaskTypeScanLibrary], option.Uid, TaskStatusNameMapping[TaskStatusRunning]),
		TaskOutput: output,
		Library:    &library,
		Option:     &option,
	}
	task.logger = TaskLogger.WithFields(logrus.Fields{
		"id":        task.Id,
		"path":      library.Path,
		"libraryId": library.ID,
	})
	module.TaskModule.Pool.AddTask(task)
	return task, nil
}

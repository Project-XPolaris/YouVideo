package task

import (
	"errors"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type ScanTask struct {
	task.BaseTask
	TaskOutput *ScanTaskOutput
	Library    database.Library
	logger     *logrus.Entry
	Option     *CreateScanTaskOption
}

func (t *ScanTask) Stop() error {
	return nil
}

func (t *ScanTask) Start() error {
	t.logger.Info("task start")
	// lock library
	if !service.DefaultLibraryLockManager.TryToLock(t.Library.ID) {
		t.logger.Error("library is busy")
	}
	defer service.DefaultLibraryLockManager.UnlockLibrary(t.Library.ID)
	// remove file where is not found
	err := service.CheckLibrary(t.Library.ID)
	if err != nil {
		t.Err = err
		if t.Option.OnError != nil {
			t.Option.OnError(t, err)
		}
		return nil
	}
	pathList, err := service.ScanVideo(&t.Library, t.Option.ExcludeDirList)
	if err != nil {
		t.Err = err
		if t.Option.OnError != nil {
			t.Option.OnError(t, err)
		}
		return nil
	}
	t.TaskOutput.Total = int64(len(pathList))
	for idx, path := range pathList {
		t.TaskOutput.Current = int64(idx + 1)
		t.TaskOutput.CurrentPath = path
		t.TaskOutput.CurrentName = filepath.Base(path)
		video, err := service.CreateVideoFile(path, t.Library.ID, t.Library.DefaultVideoType, t.Option.MatchSubject)
		if err != nil {
			if t.Option.OnFileError != nil {
				t.Option.OnFileError(t, err)
			}
			t.logger.Error(err)
			continue
		}

		if t.Option.DirectoryMode {
			parentDirName := filepath.Base(filepath.Dir(path))
			// create entity
			isCreate, entity, err := service.GetOrCreateEntityWithDirPath(parentDirName, t.Library.ID, filepath.Dir(path))
			if err != nil {
				if t.Option.OnFileError != nil {
					t.Option.OnFileError(t, err)
				}
				t.logger.Error(err)
				continue
			}
			err = service.AddVideoToEntity([]uint{video.ID}, entity.ID)
			if err != nil {
				if t.Option.OnFileError != nil {
					t.Option.OnFileError(t, err)
				}
				t.logger.Error(err)
				continue
			}
			if isCreate {
				t.logger.Info("scan meta entity ", entity.Name)
				dirItems, err := os.ReadDir(filepath.Dir(path))
				if err != nil {
					if t.Option.OnFileError != nil {
						t.Option.OnFileError(t, err)
					}
					t.logger.Error(err)
					continue
				}
				metaPath := ""
				for _, item := range dirItems {
					if item.Name() == "meta.json" {
						metaPath = filepath.Join(filepath.Dir(path), item.Name())
						break
					}
				}
				if len(metaPath) != 0 {
					err = service.ParseEntityMetaFile(entity, metaPath)
					if err != nil {
						if t.Option.OnFileError != nil {
							t.Option.OnFileError(t, err)
						}
						t.logger.Error(err)
						continue
					}
					err = database.Instance.Save(entity).Error
					if err != nil {
						if t.Option.OnFileError != nil {
							t.Option.OnFileError(t, err)
						}
						t.logger.Error(err)
						continue
					}
				}
			}
		}
		if t.Option.OnFileComplete != nil {
			t.Option.OnFileComplete(t)
		}

	}
	if service.DefaultMeilisearchEngine.Enable {
		service.DefaultMeilisearchEngine.Sync(t.Library.ID)

	}
	t.BaseTask.Status = TaskStatusNameMapping[TaskStatusDone]
	if t.Option.OnComplete != nil {
		t.Option.OnComplete(t)
	}
	return nil
}

func (t *ScanTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

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
	DirectoryMode  bool
	ExcludeDirList []string
	OnFileComplete func(task *ScanTask)
	OnFileError    func(task *ScanTask, err error)
	OnError        func(task *ScanTask, err error)
	OnComplete     func(task *ScanTask)
}

func CreateSyncLibraryTask(option CreateScanTaskOption) (*ScanTask, error) {
	existRunningTask := module.TaskModule.Pool.GetTaskWithStatus(TaskTypeNameMapping[TaskTypeScanLibrary], TaskStatusNameMapping[TaskStatusRunning])
	if existRunningTask != nil {
		return existRunningTask.(*ScanTask), nil
	}
	if service.DefaultLibraryLockManager.IsLock(option.LibraryId) {
		return nil, errors.New("library is busy")
	}
	var library database.Library
	err := database.Instance.Preload("Users").Find(&library, option.LibraryId).Error
	if err != nil {
		return nil, err
	}
	output := &ScanTaskOutput{
		Id:   library.ID,
		Path: library.Path,
	}
	if option.ExcludeDirList == nil {
		option.ExcludeDirList = []string{}
	}
	task := &ScanTask{
		BaseTask:   *task.NewBaseTask(TaskTypeNameMapping[TaskTypeScanLibrary], option.Uid, TaskStatusNameMapping[TaskStatusRunning]),
		TaskOutput: output,
		Library:    library,
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

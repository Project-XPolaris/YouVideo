package service

import (
	. "github.com/ahmetb/go-linq/v3"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"sync"
)

type Signal struct {
}

const (
	TaskTypeScanLibrary = iota + 1
)
const (
	TaskStatusRunning = iota + 1
	TaskStatusDone
	TaskStatusError
)

var TaskLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "Task",
})

var TaskTypeNameMapping map[int]string = map[int]string{
	TaskTypeScanLibrary: "ScanLibrary",
}

var TaskStatusNameMapping map[int]string = map[int]string{
	TaskStatusRunning: "Running",
	TaskStatusDone:    "Done",
	TaskStatusError:   "Error",
}
var DefaultTaskPool TaskPool = TaskPool{
	Tasks: []*Task{},
}

type Task struct {
	Id       string
	Type     int
	Status   int
	DoneChan chan Signal
	Output   interface{}
}
type TaskPool struct {
	sync.Mutex
	Tasks []*Task
}

type ScanTaskOutput struct {
	Id   uint   `json:"id"`
	Path string `json:"path"`
}

func (p *TaskPool) RemoveTaskById(id string) {
	p.Lock()
	defer p.Unlock()
	var newTask []*Task
	From(p.Tasks).Where(func(task interface{}) bool {
		return task.(*Task).Id != id
	}).ToSlice(&newTask)
	p.Tasks = newTask
}
func CreateSyncLibraryTask(library *database.Library) *Task {
	for _, task := range DefaultTaskPool.Tasks {
		if task.Output.(ScanTaskOutput).Id == library.ID {
			if task.Status == TaskStatusRunning {
				return task
			}
			DefaultTaskPool.RemoveTaskById(task.Id)
			break
		}
	}
	id := xid.New().String()
	task := &Task{
		Id:     id,
		Type:   TaskTypeScanLibrary,
		Status: TaskStatusRunning,
		Output: ScanTaskOutput{
			Id:   library.ID,
			Path: library.Path,
		},
	}
	logger := TaskLogger.WithFields(logrus.Fields{
		"id":        id,
		"path":      library.Path,
		"libraryId": library.ID,
	})
	go func() {
		logger.Info("task start")
		err := ScanLibrary(library)
		DefaultTaskPool.Lock()
		if err != nil {
			logger.Error(err)
			task.Status = TaskStatusError
		}
		task.Status = TaskStatusDone
		DefaultTaskPool.Unlock()

	}()
	DefaultTaskPool.Lock()
	DefaultTaskPool.Tasks = append(DefaultTaskPool.Tasks, task)
	DefaultTaskPool.Unlock()
	return task
}

func GetTaskList() []*Task {
	return DefaultTaskPool.Tasks
}

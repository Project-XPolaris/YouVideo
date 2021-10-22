package service

import (
	. "github.com/ahmetb/go-linq/v3"
	"github.com/sirupsen/logrus"
	"sync"
)

type Signal struct {
}

const (
	TaskTypeScanLibrary = iota + 1
	TaskTypeMeta
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
	TaskTypeMeta:        "Meta",
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
	Uid      string
}
type TaskPool struct {
	sync.Mutex
	Tasks []*Task
}

func (t *Task) SetError(err error) {
	TaskLogger.Error(err)
	if err != nil {
		t.Status = TaskStatusError
	}
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

func GetTaskList() []*Task {
	return DefaultTaskPool.Tasks
}

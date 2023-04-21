package task

import (
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/util"
)

type MD5Task struct {
	task.BaseTask
	TaskOutput *MD5Output
	Option     *MD5Option
	CheckSum   string
}

func (t *MD5Task) Stop() error {
	return nil
}

func (t *MD5Task) Start() error {
	var err error
	t.CheckSum, err = util.MD5Checksum(t.Option.FilePath)
	if err != nil {
		return t.AbortError(err)
	}
	t.Done()
	return nil
}

func (t *MD5Task) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

type MD5Output struct {
	Id   uint   `json:"id"`
	Path string `json:"path"`
}
type MD5Option struct {
	Uid          string
	FilePath     string
	ParentTaskId string
}

func CreateMD5Task(option MD5Option) *MD5Task {
	task := &MD5Task{
		BaseTask: *task.NewBaseTask(TaskTypeNameMapping[TaskMD5], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		Option:   &option,
	}
	task.ParentTaskId = option.ParentTaskId
	return task
}

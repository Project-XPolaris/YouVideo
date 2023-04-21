package task

import (
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
)

type ParseEntityMetaTaskOption struct {
	Uid          string
	Entity       *database.Entity
	MetaPath     string
	ParentTaskId string
}
type ParseEntityMetaTaskOutput struct {
}
type ParseEntityMetaTask struct {
	*task.BaseTask
	option     *ParseEntityMetaTaskOption
	TaskOutput *ParseEntityMetaTaskOutput
}

func (t *ParseEntityMetaTask) Stop() error {
	return nil
}

func (t *ParseEntityMetaTask) Start() error {
	err := service.ParseEntityMetaFile(t.option.Entity, t.option.MetaPath)
	if err != nil {
		return t.AbortError(err)
	}
	t.Done()
	return nil
}

func (t *ParseEntityMetaTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func NewParseEntityMetaTask(option *ParseEntityMetaTaskOption) *ParseEntityMetaTask {
	t := &ParseEntityMetaTask{
		BaseTask:   task.NewBaseTask(TaskTypeNameMapping[TaskParseEntityMeta], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		TaskOutput: &ParseEntityMetaTaskOutput{},
		option:     option,
	}
	t.ParentTaskId = option.ParentTaskId
	return t
}

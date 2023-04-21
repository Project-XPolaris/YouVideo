package task

import (
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
)

type GenerateCoverTaskOption struct {
	Uid          string
	file         *database.File
	ParentTaskId string
}
type GenerateCoverTaskOutput struct {
}
type GenerateCoverTask struct {
	*task.BaseTask
	option     *GenerateCoverTaskOption
	TaskOutput *GenerateCoverTaskOutput
}

func (t *GenerateCoverTask) Stop() error {
	return nil
}

func (t *GenerateCoverTask) Start() error {
	err := service.GenerateImageCover(t.option.file)
	if err != nil {
		return t.AbortError(err)
	}
	t.Done()
	return nil
}

func (t *GenerateCoverTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func NewGenerateCoverTask(option *GenerateCoverTaskOption) *GenerateCoverTask {
	t := &GenerateCoverTask{
		BaseTask:   task.NewBaseTask(TaskTypeNameMapping[TaskGenerateCover], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		TaskOutput: &GenerateCoverTaskOutput{},
		option:     option,
	}
	t.ParentTaskId = option.ParentTaskId
	return t
}

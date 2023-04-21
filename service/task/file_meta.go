package task

import (
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/service"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type AnalyzeFileMetaTaskOption struct {
	Uid          string
	path         string
	ParentTaskId string
}
type AnalyzeFileMetaTaskOutput struct {
}
type AnalyzeFileMetaTask struct {
	*task.BaseTask
	option     *AnalyzeFileMetaTaskOption
	TaskOutput *AnalyzeFileMetaTaskOutput
	MetaData   *ffprobe.ProbeData
}

func (t *AnalyzeFileMetaTask) Stop() error {
	return nil
}

func (t *AnalyzeFileMetaTask) Start() error {
	meta, err := service.GetVideoFileMeta(t.option.path)
	if err != nil {
		return t.AbortError(err)
	}
	t.MetaData = meta
	t.Done()
	return nil
}

func (t *AnalyzeFileMetaTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func NewAnalyzeFileMetaTask(option *AnalyzeFileMetaTaskOption) *AnalyzeFileMetaTask {
	t := &AnalyzeFileMetaTask{
		BaseTask:   task.NewBaseTask(TaskTypeNameMapping[TaskAnalyzeFileMeta], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		TaskOutput: &AnalyzeFileMetaTaskOutput{},
		option:     option,
	}
	t.ParentTaskId = option.ParentTaskId
	return t
}

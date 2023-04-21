package task

import (
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/service"
)

type ScanVideoTaskOption struct {
	Uid            string
	LibraryPath    string
	ExcludeDirList []string
	ParentTaskId   string
}
type ScanVideoTaskOutput struct {
}
type ScanVideoTask struct {
	*task.BaseTask
	option     *ScanVideoTaskOption
	TaskOutput *ScanVideoTaskOutput
	PathList   []string
}

func (t *ScanVideoTask) Stop() error {
	return nil
}

func (t *ScanVideoTask) Start() error {
	pathList, err := service.ScanVideo(t.option.LibraryPath, t.option.ExcludeDirList)
	if err != nil {
		return t.AbortError(err)
	}
	t.PathList = pathList
	t.Done()
	return nil
}

func (t *ScanVideoTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func NewScanVideoTask(option *ScanVideoTaskOption) *ScanVideoTask {
	t := &ScanVideoTask{
		BaseTask:   task.NewBaseTask(TaskTypeNameMapping[TaskTypeScanLibrary], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		PathList:   []string{},
		TaskOutput: &ScanVideoTaskOutput{},
		option:     option,
	}
	t.ParentTaskId = option.ParentTaskId
	return t
}

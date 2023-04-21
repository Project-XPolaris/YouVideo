package task

import (
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/util"
)

type RemoveNotExistVideoTaskOption struct {
	Uid          string
	libraryId    uint
	ParentTaskId string
}
type RemoveNotExistVideoTaskOutput struct {
}
type RemoveNotExistVideoTask struct {
	*task.BaseTask
	option     *RemoveNotExistVideoTaskOption
	TaskOutput *RemoveNotExistVideoTaskOutput
}

func (t *RemoveNotExistVideoTask) Stop() error {
	return nil
}

func (t *RemoveNotExistVideoTask) Start() error {
	library, err := service.GetLibraryById(t.option.libraryId, "Videos.Files")
	if err != nil {
		return err
	}
	for _, video := range library.Videos {
		removeCount := 0
		for _, file := range video.Files {
			if !util.CheckFileExist(file.Path) {
				err = service.RemoveFileById(file.ID)
				if err != nil {
					return t.AbortError(err)
				}
				removeCount++
			}
		}
		if removeCount == len(video.Files) {
			err = service.DeleteVideoById(video.ID)
			if err != nil {
				return t.AbortError(err)
			}
		}
	}
	t.Done()
	return nil
}

func (t *RemoveNotExistVideoTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func NewRemoveNotExistVideoTask(option *RemoveNotExistVideoTaskOption) *RemoveNotExistVideoTask {
	t := &RemoveNotExistVideoTask{
		BaseTask:   task.NewBaseTask(TaskTypeNameMapping[TaskRemoveNotExistVideo], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		TaskOutput: &RemoveNotExistVideoTaskOutput{},
		option:     option,
	}
	t.ParentTaskId = option.ParentTaskId
	return t
}

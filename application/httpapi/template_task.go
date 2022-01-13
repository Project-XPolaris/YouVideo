package httpapi

import (
	"github.com/projectxpolaris/youvideo/service/task"
)

type TaskTemplate struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string
	Output interface{}
}

func NewTaskTemplate(task *task.Task) *TaskTemplate {
	output := &TaskTemplate{}
	output.SerializeTaskTemplate(task)
	return output
}
func (t *TaskTemplate) SerializeTaskTemplate(taskData *task.Task) {
	t.Id = taskData.Id
	t.Type = task.TaskTypeNameMapping[taskData.Type]
	t.Status = task.TaskStatusNameMapping[taskData.Status]
	switch taskData.Output.(type) {
	case *task.GenerateVideoMetaTaskOutput:
		outputTemplate := ReadMetaTaskTemplate{}
		outputTemplate.Serialize(taskData.Output.(*task.GenerateVideoMetaTaskOutput))
		t.Output = outputTemplate
	default:
		t.Output = taskData.Output
	}
}

type ReadMetaTaskTemplate struct {
	LibraryId   uint   `json:"id"`
	Total       int64  `json:"total"`
	Current     int64  `json:"current"`
	CurrentPath string `json:"currentPath"`
	CurrentName string `json:"currentName"`
}

func (t *ReadMetaTaskTemplate) Serialize(output *task.GenerateVideoMetaTaskOutput) {
	t.LibraryId = output.Id
	t.Total = output.Total
	t.CurrentName = output.CurrentName
	t.Current = output.Current
	t.CurrentPath = output.CurrentPath
}

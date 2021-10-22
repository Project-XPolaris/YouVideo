package application

import "github.com/projectxpolaris/youvideo/service"

type TaskTemplate struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string
	Output interface{}
}

func NewTaskTemplate(task *service.Task) *TaskTemplate {
	output := &TaskTemplate{}
	output.SerializeTaskTemplate(task)
	return output
}
func (t *TaskTemplate) SerializeTaskTemplate(task *service.Task) {
	t.Id = task.Id
	t.Type = service.TaskTypeNameMapping[task.Type]
	t.Status = service.TaskStatusNameMapping[task.Status]
	switch task.Output.(type) {
	case *service.GenerateVideoMetaTaskOutput:
		outputTemplate := ReadMetaTaskTemplate{}
		outputTemplate.Serialize(task.Output.(*service.GenerateVideoMetaTaskOutput))
		t.Output = outputTemplate
	default:
		t.Output = task.Output
	}
}

type ReadMetaTaskTemplate struct {
	LibraryId   uint   `json:"id"`
	Total       int64  `json:"total"`
	Current     int64  `json:"current"`
	CurrentPath string `json:"currentPath"`
	CurrentName string `json:"currentName"`
}

func (t *ReadMetaTaskTemplate) Serialize(output *service.GenerateVideoMetaTaskOutput) {
	t.LibraryId = output.Id
	t.Total = output.Total
	t.CurrentName = output.CurrentName
	t.Current = output.Current
	t.CurrentPath = output.CurrentPath
}

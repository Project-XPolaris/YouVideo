package httpapi

import (
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youvideo/module"
	taskService "github.com/projectxpolaris/youvideo/service/task"
)

//type TaskTemplate struct {
//	Id     string `json:"id"`
//	Type   string `json:"type"`
//	Status string
//	Output interface{}
//}
//
func NewTaskTemplate(task task.Task) interface{} {
	tp, _ := module.TaskModule.SerializerTemplate(task)
	return tp
}

//func (t *TaskTemplate) SerializeTaskTemplate(taskData task.Task) {
//	t.Id = taskData.GetId()
//	t.Type = taskData.GetType()
//	t.Status = taskData.GetStatus()
//	output, _ := taskData.Output()
//	switch output.(type) {
//	case *taskService.GenerateVideoMetaTaskOutput:
//		outputTemplate := ReadMetaTaskTemplate{}
//		outputTemplate.Serialize(output.(*taskService.GenerateVideoMetaTaskOutput))
//		t.Output = outputTemplate
//	default:
//		t.Output = output
//	}
//}
func NewReadMetaTaskTemplate(data *taskService.GenerateVideoMetaTaskOutput) (*ReadMetaTaskTemplate, error) {
	template := &ReadMetaTaskTemplate{}
	template.Serialize(data)
	return template, nil
}

type ReadMetaTaskTemplate struct {
	LibraryId   uint   `json:"id"`
	Total       int64  `json:"total"`
	Current     int64  `json:"current"`
	CurrentPath string `json:"currentPath"`
	CurrentName string `json:"currentName"`
}

func (t *ReadMetaTaskTemplate) Serialize(output *taskService.GenerateVideoMetaTaskOutput) {
	t.LibraryId = output.Id
	t.Total = output.Total
	t.CurrentName = output.CurrentName
	t.Current = output.Current
	t.CurrentPath = output.CurrentPath
}

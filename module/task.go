package module

import "github.com/allentom/harukap/module/task"

var TaskModule *task.TaskModule

func CreateTaskModule() {
	TaskModule = task.NewTaskModule()
}

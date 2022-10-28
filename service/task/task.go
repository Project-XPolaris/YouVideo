package task

import (
	"github.com/sirupsen/logrus"
)

type Signal struct {
}

const (
	TaskTypeScanLibrary = iota + 1
	TaskTypeMeta
	TaskTypeRemove
	TaskTypeMatchEntity
	TaskTypeSyncIndex
)
const (
	TaskStatusRunning = iota + 1
	TaskStatusDone
	TaskStatusError
)

var TaskLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "Task",
})

var TaskTypeNameMapping map[int]string = map[int]string{
	TaskTypeScanLibrary: "ScanLibrary",
	TaskTypeMeta:        "Meta",
	TaskTypeRemove:      "RemoveLibrary",
	TaskTypeMatchEntity: "MatchEntity",
	TaskTypeSyncIndex:   "SyncIndex",
}

var TaskStatusNameMapping map[int]string = map[int]string{
	TaskStatusRunning: "Running",
	TaskStatusDone:    "Done",
	TaskStatusError:   "Error",
}

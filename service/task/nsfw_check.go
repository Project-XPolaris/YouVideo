package task

import (
	"github.com/allentom/harukap/module/task"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
	"io"
)

type NSFWCheckResult struct {
	Hentai bool
	Sexy   bool
	Porn   bool
}
type NSFWCheckTaskOption struct {
	Uid          string
	path         string
	ParentTaskId string
}
type NSFWCheckTaskOutput struct {
}
type NSFWCheckTask struct {
	*task.BaseTask
	option     *NSFWCheckTaskOption
	TaskOutput *NSFWCheckTaskOutput
	Result     *NSFWCheckResult
}

func (t *NSFWCheckTask) Stop() error {
	return nil
}

func (t *NSFWCheckTask) Start() error {
	taskLogger := plugin.DefaultYouLogPlugin.Logger.NewScope("nsfw check task")
	outChan := make(chan io.Reader)
	go func() {
		err := util.ExtractNShotFromVideoPipe(t.option.path, config.Instance.NSFWCheckConfig.Slice, outChan)
		if err != nil {
			taskLogger.WithFields(youlog.Fields{
				"error": err,
			}).Error("extract frame error")
		}
	}()

	for shot := range outChan {
		result, err := plugin.DefaultNSFWCheckPlugin.Client.Predict(shot)
		if err != nil {
			taskLogger.WithFields(youlog.Fields{
				"error": err,
			}).Error("nsfw check error")
			continue
		}
		for _, predictions := range result {
			if predictions.Probability > 0.5 {
				switch predictions.Classname {
				case "Hentai":
					t.Result.Hentai = true
					break
				case "Porn":
					t.Result.Porn = true
					break
				case "Sexy":
					t.Result.Sexy = true
					break
				}
			}
			if t.Result.Hentai || t.Result.Porn || t.Result.Sexy {
				break
			}
		}
	}
	t.Done()
	return nil
}

func (t *NSFWCheckTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func NewNSFWCheckTask(option *NSFWCheckTaskOption) *NSFWCheckTask {
	t := &NSFWCheckTask{
		BaseTask:   task.NewBaseTask(TaskTypeNameMapping[TaskNSFWCheck], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		TaskOutput: &NSFWCheckTaskOutput{},
		Result:     &NSFWCheckResult{},
		option:     option,
	}
	t.ParentTaskId = option.ParentTaskId
	return t
}

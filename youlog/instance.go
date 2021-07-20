package youlog

import (
	logtoolkit "github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/config"
)

var DefaultClient *logtoolkit.LogClient = &logtoolkit.LogClient{}

func Init() {
	DefaultClient.Init(config.Instance.Addr, config.Instance.Application, config.Instance.Instance)
}

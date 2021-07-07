package youplus

import (
	"errors"
	"github.com/project-xpolaris/youplustoolkit"
	"github.com/projectxpolaris/youvideo/config"
)

var DefaultClient *youplustoolkit.Client

func InitClient() error {
	DefaultClient = youplustoolkit.NewClient()
	DefaultClient.Init(config.Instance.YouPlusUrl)
	info, err := DefaultClient.GetInfo()
	if err != nil {
		return err
	}
	if !info.Success {
		return errors.New("get info not successful")
	}
	return nil
}

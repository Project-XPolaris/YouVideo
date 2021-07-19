package youplus

import (
	"errors"
	plustoolkit "github.com/project-xpolaris/youplustoolkit/youplus"
	"github.com/projectxpolaris/youvideo/config"
)

var DefaultClient *plustoolkit.Client

func InitClient() error {
	DefaultClient = plustoolkit.NewClient()
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

package plugin

import (
	"fmt"
	"github.com/allentom/harukap"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/util"
	"os"
)

type InitPlugin struct {
}

func (p *InitPlugin) OnInit(e *harukap.HarukaAppEngine) error {
	logger := e.LoggerPlugin.Logger.NewScope("InitPlugin")
	configManager := e.ConfigProvider.Manager
	configManager.GetString("temp_store")
	if !util.CheckFileExist(config.Instance.TempStore) {
		err := os.Mkdir(config.Instance.TempStore, 0777)
		if err != nil {
			return err
		}
	}
	// check library
	var libraryList []database.Library
	database.Instance.Find(&libraryList)
	for _, library := range libraryList {
		if !util.CheckFileExist(library.Path) {
			logger.Warn(fmt.Sprintf("Library [%s] not exist", library.Path))
		}
	}
	return nil
}

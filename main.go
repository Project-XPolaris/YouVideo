package main

import (
	"github.com/allentom/harukap"
	cli2 "github.com/allentom/harukap/cli"
	config2 "github.com/allentom/harukap/config"
	"github.com/projectxpolaris/youvideo/application"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/youlog"
	"github.com/projectxpolaris/youvideo/youplus"
	"github.com/sirupsen/logrus"
)

func main() {
	err := config.InitConfigProvider(func(provider *config2.Provider) {
		config.ReadConfig(provider)
	})
	if err != nil {
		logrus.Fatal(err)
	}
	err = youlog.DefaultYouLogPlugin.OnInit(config.DefaultConfigProvider)
	if err != nil {
		logrus.Fatal(err)
	}

	appEngine := harukap.NewHarukaAppEngine()
	appEngine.ConfigProvider = config.DefaultConfigProvider
	appEngine.LoggerPlugin = youlog.DefaultYouLogPlugin
	appEngine.UsePlugin(&youplus.DefaultYouPlusPlugin)
	appEngine.UsePlugin(database.DefaultPlugin)
	appEngine.HttpService = application.GetEngine()

	appWrap, err := cli2.NewWrapper(appEngine)
	if err != nil {
		logrus.Fatal(err)
	}
	appWrap.RunApp()
}

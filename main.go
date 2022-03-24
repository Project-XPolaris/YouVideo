package main

import (
	"context"
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/cli"
	"github.com/allentom/harukap/thumbnail"
	"github.com/projectxpolaris/youvideo/application/httpapi"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/youlog"
	"github.com/projectxpolaris/youvideo/youplus"
	"github.com/sirupsen/logrus"
)

func main() {
	err := config.InitConfigProvider()
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
	if config.Instance.ThumbnailType == "thumbnailservice" {
		plugin.DefaultThumbnailPlugin.SetConfig(&thumbnail.ThumbnailServiceConfig{
			Enable:     true,
			ServiceUrl: config.Instance.ThumbnailServiceUrl,
		})
		appEngine.UsePlugin(plugin.DefaultThumbnailPlugin)
	}
	appEngine.UsePlugin(&plugin.DefaultRegisterPlugin)
	appEngine.HttpService = httpapi.GetEngine()
	if err != nil {
		logrus.Fatal(err)
	}
	if config.Instance.YouLibraryConfig.Enable {
		service.DefaultVideoInformationMatchService.Init()
		go service.DefaultVideoInformationMatchService.Run(context.Background())
	}
	appWrap, err := cli.NewWrapper(appEngine)
	if err != nil {
		logrus.Fatal(err)
	}
	appWrap.RunApp()
}

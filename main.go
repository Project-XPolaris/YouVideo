package main

import (
	"context"
	"fmt"
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/cli"
	"github.com/projectxpolaris/youvideo/application/httpapi"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/sirupsen/logrus"
)

func main() {
	err := config.InitConfigProvider()
	if err != nil {
		logrus.Fatal(err)
	}
	err = plugin.DefaultYouLogPlugin.OnInit(config.DefaultConfigProvider)
	if err != nil {
		logrus.Fatal(err)
	}

	appEngine := harukap.NewHarukaAppEngine()
	appEngine.ConfigProvider = config.DefaultConfigProvider
	appEngine.LoggerPlugin = plugin.DefaultYouLogPlugin
	plugin.CreateDefaultYouPlusPlugin()
	appEngine.UsePlugin(plugin.DefaultYouPlusPlugin)
	appEngine.UsePlugin(database.DefaultPlugin)
	appEngine.UsePlugin(&plugin.DefaultRegisterPlugin)
	appEngine.UsePlugin(&plugin.InitPlugin{})
	appEngine.UsePlugin(plugin.StorageEnginePlugin)
	appEngine.UsePlugin(plugin.DefaultThumbnailPlugin)
	// init auth
	rawAuth := config.DefaultConfigProvider.Manager.GetStringMap("auth")
	for key, _ := range rawAuth {
		rawAuthContent := config.DefaultConfigProvider.Manager.GetString(fmt.Sprintf("auth.%s.type", key))
		if rawAuthContent == "youauth" {
			plugin.CreateYouAuthPlugin()
			plugin.DefaultYouAuthOauthPlugin.ConfigPrefix = fmt.Sprintf("auth.%s", key)
			appEngine.UsePlugin(plugin.DefaultYouAuthOauthPlugin)
		}
	}
	module.CreateAuthModule()
	err = module.CreateNotificationModule()
	if err != nil {
		logrus.Fatal(err)
	}
	module.CreateTaskModule()
	appEngine.HttpService = httpapi.GetEngine()
	if err != nil {
		logrus.Fatal(err)
	}
	service.InitTMDB()
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

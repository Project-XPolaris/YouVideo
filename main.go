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
	logger := plugin.DefaultYouLogPlugin.Logger.NewScope("main")
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
	appEngine.UsePlugin(plugin.DefaultMeiliSearchPlugin)
	appEngine.UsePlugin(plugin.DefaultNSFWCheckPlugin)
	// init auth
	rawAuth := config.DefaultConfigProvider.Manager.GetStringMap("auth")
	for key, _ := range rawAuth {
		logger.Info(fmt.Sprintf("init auth %s", key))
		rawAuthContent := config.DefaultConfigProvider.Manager.GetString(fmt.Sprintf("auth.%s.type", key))
		if rawAuthContent == "youauth" {
			plugin.CreateYouAuthPlugin()
			plugin.DefaultYouAuthOauthPlugin.ConfigPrefix = fmt.Sprintf("auth.%s", key)
			appEngine.UsePlugin(plugin.DefaultYouAuthOauthPlugin)
		}
	}
	logger.Info("init auth module")
	module.CreateAuthModule()
	err = module.CreateNotificationModule()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("init auth complete")
	module.CreateTaskModule()
	appEngine.HttpService = httpapi.GetEngine()
	if err != nil {
		logger.Fatal(err)
	}
	service.InitTMDB()
	service.InitBangumiInfoSource()
	appEngine.OnPluginInitComplete = func() {
		err = service.InitMeiliSearch()
		if err != nil {
			logger.Fatal(err)
		}
	}
	logger.Info("init ffmpeg and ffprobe")
	err = service.InitCheckFfmpeg()
	if err != nil {
		logger.Fatal(err)
	}
	if config.Instance.YouLibraryConfig.Enable {
		service.DefaultVideoInformationMatchService.Init()
		go service.DefaultVideoInformationMatchService.Run(context.Background())
	}
	appWrap, err := cli.NewWrapper(appEngine)
	if err != nil {
		logger.Fatal(err)
	}
	appWrap.RunApp()
}

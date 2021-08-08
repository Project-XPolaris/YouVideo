package main

import (
	srv "github.com/kardianos/service"
	logtoolkit "github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/application"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/youlog"
	"github.com/projectxpolaris/youvideo/youplus"
	"github.com/projectxpolaris/youvideo/youtrans"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
)

var svcConfig *srv.Config
var Logger = logrus.WithField("scope", "main")

func initService() error {
	workPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	svcConfig = &srv.Config{
		Name:             "YouVideoService",
		DisplayName:      "YouVideo Service",
		WorkingDirectory: workPath,
		Arguments:        []string{"run"},
	}
	return nil
}
func Program() {
	err := config.ReadConfig()
	if err != nil {
		Logger.Fatal(err)
	}
	youlog.Init()
	if config.Instance.YouLogEnable {
		err = youlog.DefaultClient.Connect()
		if err != nil {
			Logger.Fatal(err)
		}
	}
	logScope := youlog.DefaultClient.NewScope("booting")
	logScope.Info("booting application")
	if config.Instance.EnableTranscode {
		logScope.Info("check transcode [checking]")
		_, err = youtrans.DefaultYouTransClient.GetInfo()
		if err != nil {
			logScope.WithFields(logtoolkit.Fields{
				"url": config.Instance.YoutransURL,
			}).Fatal(err.Error())
		}
		logScope.WithFields(logtoolkit.Fields{
			"url": config.Instance.YoutransURL,
		}).Info("check transcode [pass]")
	}
	// youplus enable
	if config.Instance.YouPlusPath || config.Instance.EnableAuth {
		logScope.Info("check youplus [checking]")
		err = youplus.InitClient()
		if err != nil {
			logScope.WithFields(logtoolkit.Fields{
				"url": config.Instance.YouPlusUrl,
			}).Fatal(err.Error())
		}
		logScope.WithFields(logtoolkit.Fields{
			"url": config.Instance.YoutransURL,
		}).Info("check youplus service [pass]")
	}
	logScope.Info("booting success")
	application.Run()
}

type program struct{}

func (p *program) Start(s srv.Service) error {
	go Program()
	return nil
}

func (p *program) Stop(s srv.Service) error {
	return nil
}

func InstallAsService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()

	err = s.Install()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful install service")
}

func UnInstall() {

	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful uninstall service")
}

func StartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
func StopService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Stop()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RestartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Restart()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RunApp() {
	app := &cli.App{
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "service",
				Usage: "service manager",
				Subcommands: []*cli.Command{
					{
						Name:  "install",
						Usage: "install service",
						Action: func(context *cli.Context) error {
							InstallAsService()
							return nil
						},
					},
					{
						Name:  "uninstall",
						Usage: "uninstall service",
						Action: func(context *cli.Context) error {
							UnInstall()
							return nil
						},
					},
					{
						Name:  "start",
						Usage: "start service",
						Action: func(context *cli.Context) error {
							StartService()
							return nil
						},
					},
					{
						Name:  "stop",
						Usage: "stop service",
						Action: func(context *cli.Context) error {
							StopService()
							return nil
						},
					},
					{
						Name:  "restart",
						Usage: "restart service",
						Action: func(context *cli.Context) error {
							RestartService()
							return nil
						},
					},
				},
				Description: "YouVideo service controller",
			},
			{
				Name:  "run",
				Usage: "run app",
				Action: func(context *cli.Context) error {
					Program()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := initService()
	if err != nil {
		logrus.Fatal(err)
	}
	RunApp()
}

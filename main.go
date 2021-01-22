package main

import (
	"github.com/projectxpolaris/youvideo/application"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/sirupsen/logrus"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	application.Run()
}

package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var Logger = log.New().WithFields(log.Fields{
	"scope": "Application",
})

func Run() {
	err := database.Connect()
	if err != nil {
		return
	}
	e := haruka.NewEngine()
	e.UseCors(cors.AllowAll())
	e.UseMiddleware(middleware.NewLoggerMiddleware())
	e.UseMiddleware(middleware.NewPaginationMiddleware("page", "pageSize", 1, 20))
	e.Router.POST("/library", createLibraryHandler)
	e.Router.GET("/library", readLibraryList)
	e.Router.POST("/library/{id:[0-9]+}/scan", scanLibrary)
	e.Router.DELETE("/library/{id:[0-9]+}", deleteLibrary)
	e.Router.GET("/videos", readVideoList)
	e.Router.DELETE("/video/{id:[0-9]+}", deleteVideoHandler)
	e.Router.GET("/video/{id:[0-9]+}/stream", playVideo)
	e.Router.POST("/video/{id:[0-9]+}/move", moveVideoHandler)
	e.Router.GET("/files", readDirectoryHandler)
	e.Router.GET("/task", readTaskListHandler)
	e.Router.Static("/covers", config.AppConfig.CoversStore)
	Logger.Info("application started")
	e.RunAndListen(config.AppConfig.Addr)
}

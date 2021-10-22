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
	err = database.InitDatabase()
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
	e.Router.POST("/library/{id:[0-9]+}/meta", readMetaTask)
	e.Router.DELETE("/library/{id:[0-9]+}", deleteLibrary)
	e.Router.GET("/videos", readVideoList)
	e.Router.DELETE("/video/{id:[0-9]+}", deleteVideoHandler)
	e.Router.GET("/video/{id:[0-9]+}", getVideoHandler)
	e.Router.GET("/video/file/{id:[0-9]+}/stream", playVideo)
	e.Router.GET("/video/file/{id:[0-9]+}/cover", fileCoverHandler)
	e.Router.GET("/video/file/{id:[0-9]+}/subtitles", videoSubtitle)
	e.Router.POST("/video/{id:[0-9]+}/move", moveVideoHandler)
	e.Router.POST("/video/{id:[0-9]+}/trans", transcodeHandler)
	e.Router.POST("/tag", createTagHandler)
	e.Router.GET("/tag", getTagListHandler)
	e.Router.DELETE("/tag/{id:[0-9]+}", removeTagHandler)
	e.Router.PATCH("/tag/{id:[0-9]+}", updateTagHandler)
	e.Router.POST("/tag/{id:[0-9]+}/videos", addVideoToTagHandler)
	e.Router.DELETE("/tag/{id:[0-9]+}/videos", removeVideoFromTagHandler)
	e.Router.POST("/tag/videos", tagVideosBatchHandler)
	e.Router.GET("/ffmpeg/codec", getCodecsHandler)
	e.Router.GET("/ffmpeg/formats", getFormatsHandler)
	e.Router.GET("/files", readDirectoryHandler)
	e.Router.GET("/task", readTaskListHandler)
	e.Router.Static("/covers", config.Instance.CoversStore)
	e.Router.GET("/info", serviceInfoHandler)
	e.Router.DELETE("/file/{id:[0-9]+}", removeFileHandler)
	e.Router.POST("/file/{id:[0-9]+}/rename", renameFileHandler)
	e.Router.POST("/callback/tran/complete", transCompleteCallback)
	e.Router.GET("/history", getHistoryListHandler)
	e.Router.GET("/user/auth", youPlusTokenHandler)
	e.Router.POST("/user/auth", youPlusLoginHandler)
	e.Router.AddHandler("/notification", notificationSocketHandler)
	e.UseMiddleware(&AuthMiddleware{})
	e.UseMiddleware(&ReadUserMiddleware{})
	Logger.Info("application started")
	e.RunAndListen(config.Instance.Addr)
}

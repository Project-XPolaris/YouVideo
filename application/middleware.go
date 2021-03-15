package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/auth"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/sirupsen/logrus"
	"strings"
)

var noAuthPath = []string{}

type AuthMiddleware struct {
}

func (a *AuthMiddleware) OnRequest(ctx *haruka.Context) {
	if !config.Instance.EnableAuth {
		ctx.Param["uid"] = service.PublicUid
		ctx.Param["username"] = service.PublicUsername
		return
	}
	for _, targetPath := range noAuthPath {
		if ctx.Request.URL.Path == targetPath {
			return
		}
	}
	rawString := ctx.Request.Header.Get("Authorization")
	if len(rawString) > 0 {
		rawString = strings.Replace(rawString, "Bearer ", "", 1)
		response, err := auth.DefaultAuthClient.CheckAuth(rawString)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"uid":  response.Uid,
				"user": response.Username,
			}).Info("user auth")
			ctx.Param["uid"] = response.Uid
			ctx.Param["username"] = response.Username
		} else {
			ctx.Param["uid"] = service.PublicUid
			ctx.Param["username"] = service.PublicUsername
		}
	}
}

type ReadUserMiddleware struct {
}

func (m *ReadUserMiddleware) OnRequest(ctx *haruka.Context) {
	user, _ := service.GetUserById(ctx.Param["uid"].(string))
	ctx.Param["user"] = user
}

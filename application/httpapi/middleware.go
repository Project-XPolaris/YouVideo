package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
	"strings"
)

var noAuthPath = []string{}

type AuthMiddleware struct {
}

func (a *AuthMiddleware) OnRequest(ctx *haruka.Context) {
	if !config.Instance.EnableAuth {
		ctx.Param["uid"] = service.PublicUid
		ctx.Param["username"] = service.PublicUsername
		ctx.Param["token"] = ""
		return
	}
	for _, targetPath := range noAuthPath {
		if ctx.Request.URL.Path == targetPath {
			return
		}
	}
	rawString := ctx.Request.Header.Get("Authorization")
	if len(rawString) == 0 {
		rawString = ctx.GetQueryString("token")
	}
	ctx.Param["token"] = rawString
	if len(rawString) > 0 {
		rawString = strings.Replace(rawString, "Bearer ", "", 1)
		response, err := plugin.DefaultYouPlusPlugin.Client.CheckAuth(rawString)
		if err == nil && response.Success {
			ctx.Param["uid"] = response.Uid
			ctx.Param["username"] = response.Username
		} else {
			ctx.Param["uid"] = service.PublicUid
			ctx.Param["username"] = service.PublicUsername
		}
	} else {
		ctx.Param["uid"] = service.PublicUid
		ctx.Param["username"] = service.PublicUsername
	}
}

type ReadUserMiddleware struct {
}

func (m *ReadUserMiddleware) OnRequest(ctx *haruka.Context) {
	user, _ := service.GetUserById(ctx.Param["uid"].(string))
	ctx.Param["user"] = user
}

type OauthMiddleware struct {
}

func (m *OauthMiddleware) OnRequest(ctx *haruka.Context) {
	if !config.Instance.EnableAuth {
		ctx.Param["uid"] = service.PublicUid
		ctx.Param["username"] = service.PublicUsername
		ctx.Param["token"] = ""
		return
	}
	for _, targetPath := range noAuthPath {
		if ctx.Request.URL.Path == targetPath {
			return
		}
	}
	rawString := ctx.Request.Header.Get("Authorization")
	if len(rawString) == 0 {
		rawString = ctx.GetQueryString("token")
	}
	if len(rawString) > 0 {
		rawString = strings.Replace(rawString, "Bearer ", "", 1)
		user, err := service.GetUserByAuthToken(rawString)
		if err != nil {
			ctx.Param["uid"] = service.PublicUid
			ctx.Param["username"] = service.PublicUsername
			ctx.Param["token"] = ""
		} else {
			ctx.Param["uid"] = user.Uid
			ctx.Param["username"] = user.Username
			ctx.Param["token"] = rawString
		}
	} else {
		ctx.Param["uid"] = service.PublicUid
		ctx.Param["username"] = service.PublicUsername
		ctx.Param["token"] = ""
	}
}

package httpapi

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
)

type CheckAuthMiddleware struct{}

func (m *CheckAuthMiddleware) OnRequest(ctx *haruka.Context) {
	if claims, ok := ctx.Param["claim"]; ok {
		user := claims.(*database.User)
		ctx.Param["user"] = user
		ctx.Param["uid"] = user.Uid
		ctx.Param["username"] = user.Username
		return
	}
	// for public user
	publicUser, err := service.GetUserById("-1")
	if err != nil {
		AbortError(ctx, errors.New("public user not found"), 500)
		return
	}
	ctx.Param["user"] = publicUser
	ctx.Param["uid"] = publicUser.Uid
	ctx.Param["username"] = publicUser.Username
}

type VideoAccessibleMiddleware struct{}

func (m *VideoAccessibleMiddleware) OnRequest(ctx *haruka.Context) {
	matchPatterns := []string{
		"/video/{id:[0-9]+}",
		"/video/{id:[0-9]+}",
		"/video/{id:[0-9]+}",
		"/video/{id:[0-9]+}/meta",
		"/video/{id:[0-9]+}/trans",
	}
	isMatch := false
	for _, pattern := range matchPatterns {
		if ctx.Pattern == pattern {
			isMatch = true
			break
		}
	}
	if !isMatch {
		return
	}
	if !checkVideoAccessibleAndRaiseError(ctx) {
		ctx.Abort()
	}
}

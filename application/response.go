package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/youlog"
)

func AbortError(ctx *haruka.Context, err error, status int) {
	youlog.DefaultClient.Error(err.Error())
	ctx.Writer.Header().Set("Content-Type", "application/json")
	ctx.Writer.WriteHeader(status)
	ctx.JSON(haruka.JSON{
		"success": false,
		"err":     err.Error(),
	})
}

package application

import "github.com/allentom/haruka"

func AbortError(ctx *haruka.Context, err error, status int) {
	ctx.Writer.WriteHeader(status)
	ctx.JSON(haruka.JSON{
		"success": false,
		"err":     err,
	})
}

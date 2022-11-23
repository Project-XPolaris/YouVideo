package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/commons"
	"github.com/projectxpolaris/youvideo/plugin"
	"net/http"
)

func AbortError(ctx *haruka.Context, err error, status int) {
	if apiError, ok := err.(*commons.APIError); ok {
		plugin.DefaultYouLogPlugin.Logger.Error(apiError.Err.Error())
		ctx.JSONWithStatus(haruka.JSON{
			"success": false,
			"err":     apiError.Desc,
			"code":    apiError.Code,
		}, status)
		return
	}
	plugin.DefaultYouLogPlugin.Logger.Error(err.Error())
	ctx.JSONWithStatus(haruka.JSON{
		"success": false,
		"err":     err.(error).Error(),
		"code":    "9999",
	}, status)
}

func BindingOrRaiseError(ctx *haruka.Context, target interface{}) bool {
	err := ctx.BindingInput(target)
	if err != nil {
		AbortError(ctx, &commons.APIError{
			Err:  err,
			Code: commons.CodeParseError,
			Desc: err.Error(),
		}, http.StatusBadRequest)
		return false
	}
	return true
}

func SendSuccessResponse(ctx *haruka.Context, data interface{}) {
	respData := haruka.JSON{
		"success": true,
	}
	if data != nil {
		respData["data"] = data
	}
	ctx.JSON(respData)
}

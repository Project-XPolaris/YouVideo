package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
)

var searchHandler haruka.RequestHandler = func(context *haruka.Context) {
	uid := context.Param["uid"].(string)
	content, err := service.SearchData(context.GetQueryString("q"), uid)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"result":  content,
	})
}

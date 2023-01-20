package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
)

type CreateHistoryResponseBody struct {
	VideoId uint `json:"videoId"`
}

var createHistoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateHistoryResponseBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	user := context.Param["user"].(*database.User)
	err = service.CreateHistory(body.VideoId, user.Uid)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, nil)
}
var getHistoryListHandler haruka.RequestHandler = func(context *haruka.Context) {
	var option service.HistoryQueryOption
	err := context.BindingInput(&option)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	count, historyList, err := service.GetHistoryList(option)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, haruka.JSON{
		"count":    count,
		"page":     option.Page,
		"pageSize": option.PageSize,
		"result":   serializer.SerializeMultipleTemplate(historyList, &BaseHistoryTemplate{}, map[string]interface{}{}),
	})
}

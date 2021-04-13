package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
)

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
	context.JSON(haruka.JSON{
		"count":    count,
		"page":     option.Page,
		"pageSize": option.PageSize,
		"result":   serializer.SerializeMultipleTemplate(historyList, &BaseHistoryTemplate{}, map[string]interface{}{}),
	})
}

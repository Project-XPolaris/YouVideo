package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/projectxpolaris/youvideo/service"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var getFolderListHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.FolderQueryBuilder{}
	if !BindingOrRaiseError(context, &queryBuilder) {
		return
	}
	count, result, err := queryBuilder.Read()
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data := serializer.SerializeMultipleTemplate(result, &BaseFolderTemplate{}, nil)
	log.Info(result)
	context.JSON(haruka.JSON{
		"success":  true,
		"count":    count,
		"page":     queryBuilder.Page,
		"pageSize": queryBuilder.PageSize,
		"result":   data,
	})
}

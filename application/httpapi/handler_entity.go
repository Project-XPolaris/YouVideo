package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/project-xpolaris/youplustoolkit/youlink"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
)

type CreateEntityRequestBody struct {
	LibraryId uint   `json:"libraryId"`
	Name      string `json:"name"`
}

var createEntityHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateEntityRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		youlink.AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	entity, err := service.CreateEntity(requestBody.Name, requestBody.LibraryId)
	if err != nil {
		youlink.AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	template := BaseEntityTemplate{}
	template.Serializer(entity, map[string]interface{}{})
	context.JSON(template)
}

var getEntitiesHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.EntityQueryBuilder{
		Page:     context.Param["page"].(int),
		PageSize: context.Param["pageSize"].(int),
	}
	err := context.BindingInput(&queryBuilder)
	if err != nil {
		youlink.AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	entities, total, err := queryBuilder.Query()
	if err != nil {
		youlink.AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success":  true,
		"count":    total,
		"page":     queryBuilder.Page,
		"pageSize": queryBuilder.PageSize,
		"result":   serializer.SerializeMultipleTemplate(entities, &BaseEntityTemplate{}, map[string]interface{}{}),
	})
}

package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
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
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	entity, err := service.CreateEntity(requestBody.Name, requestBody.LibraryId)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
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
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	entities, total, err := queryBuilder.Query()
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
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

type AddVideoToEntityRequestBody struct {
	Ids []uint `json:"ids"`
}

var addVideoToEntityHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody AddVideoToEntityRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	entityId, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.AddVideoToEntity(requestBody.Ids, uint(entityId))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})

}

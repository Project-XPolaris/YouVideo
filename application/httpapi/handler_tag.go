package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
	"strconv"
)

type CreateTagRequestBody struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

var createTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateTagRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	uid := service.PublicUid
	if requestBody.Private {
		uid = context.Param["uid"].(string)
	}
	if err = validator.RunValidators(
		&DuplicateTagValidator{Name: requestBody.Name, Uid: uid},
	); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	tag, err := service.CreateTag(requestBody.Name, uid)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := &BaseTagTemplate{}
	template.Serializer(tag, map[string]interface{}{})
	context.JSON(template)
}

var getTagListHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.TagQueryBuilder{}
	if err := context.BindingInput(&queryBuilder); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	count, models, err := queryBuilder.Query()
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data := make([]BaseTagTemplate, 0)
	for _, tag := range models {
		template := BaseTagTemplate{}
		template.Serializer(tag, map[string]interface{}{})
		data = append(data, template)
	}
	context.JSON(haruka.JSON{
		"success":  true,
		"count":    count,
		"page":     queryBuilder.Page,
		"pageSize": queryBuilder.PageSize,
		"result":   data,
	})
}
var removeTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permissionValidator := TagOwnerPermission{}
	err = context.BindingInput(&permissionValidator)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	validators := []validator.Validator{
		&permissionValidator,
	}
	if err := validator.RunValidators(validators...); err != nil {
		AbortError(context, err, http.StatusBadRequest)
	}
	tag := &database.Tag{}
	err = tag.DeleteById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var updateTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permissionValidator := TagOwnerPermission{}
	err = context.BindingInput(&permissionValidator)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	body := map[string]interface{}{}
	err = context.ParseJson(&body)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
	}
	duplicateTagValidator := DuplicateTagValidator{
		Name: body["name"].(string),
		Uid:  context.Param["uid"].(string),
	}
	validators := []validator.Validator{
		&permissionValidator,
		&duplicateTagValidator,
	}
	if err := validator.RunValidators(validators...); err != nil {
		AbortError(context, err, http.StatusBadRequest)
	}

	tagModel := &database.Tag{}
	rawTag, err := tagModel.UpdateById(uint(id), body)
	template := BaseTagTemplate{}
	template.Serializer(rawTag, map[string]interface{}{})
	context.JSON(template)
}

type TagVideoBatchRequestBody struct {
	Ids []uint
}

var addVideoToTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody TagVideoBatchRequestBody
	err = context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permissionValidator := TagOwnerPermission{}
	context.BindingInput(&permissionValidator)
	if err = validator.RunValidators(&permissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.AddVideosToTag(uint(id), requestBody.Ids...)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var removeVideoFromTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody TagVideoBatchRequestBody
	err = context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permissionValidator := TagOwnerPermission{}
	context.BindingInput(&permissionValidator)
	if err = validator.RunValidators(&permissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.RemoveVideosFromTag(uint(id), requestBody.Ids...)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type TagVideosRequestBody struct {
	Name []string `json:"name"`
	Ids  []uint   `json:"ids"`
}

var tagVideosBatchHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody TagVideosRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.AddOrCreateTagFromVideo(requestBody.Name, context.Param["uid"].(string), requestBody.Ids...)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

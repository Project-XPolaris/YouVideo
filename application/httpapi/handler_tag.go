package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/blueprint"
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
	view := blueprint.DeleteModelView{
		Context: context,
		Lookup:  "id",
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
		GetValidators: func(v *blueprint.DeleteModelView) []validator.Validator {
			permissionValidator := TagOwnerPermission{}
			context.BindingInput(&permissionValidator)
			return []validator.Validator{
				&permissionValidator,
			}
		},
		Model: &database.Tag{},
	}
	view.Run()
}

var updateTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := blueprint.UpdateModelView{
		Context: context,
		Lookup:  "id",
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
		Model:    &database.Tag{},
		Template: &BaseTagTemplate{},
		GetValidators: func(v *blueprint.UpdateModelView) []validator.Validator {
			permissionValidator := TagOwnerPermission{}
			context.BindingInput(&permissionValidator)
			duplicateTagValidator := DuplicateTagValidator{
				Name: v.RequestBody["name"].(string),
				Uid:  context.Param["uid"].(string),
			}
			return []validator.Validator{
				&permissionValidator,
				&duplicateTagValidator,
			}
		},
	}
	view.Run()
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

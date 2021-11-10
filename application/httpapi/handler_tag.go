package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/blueprint"
	"github.com/allentom/haruka/serializer"
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
	view := blueprint.CreateModelView{
		Context: context,
		CreateModel: func() interface{} {
			return &database.Tag{}
		},
		ResponseTemplate: &BaseTagTemplate{},
		RequestBody:      &CreateTagRequestBody{},
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},

		OnCreate: func(view *blueprint.CreateModelView, model interface{}) (interface{}, error) {
			body := view.RequestBody.(*CreateTagRequestBody)
			uid := service.PublicUid
			if body.Private {
				uid = context.Param["uid"].(string)
			}
			tag, err := service.CreateTag(body.Name, uid)
			return tag, err
		},
		GetValidators: func(v *blueprint.CreateModelView) []validator.Validator {
			tag := v.RequestBody.(*CreateTagRequestBody)
			uid := service.PublicUid
			if tag.Private {
				uid = context.Param["uid"].(string)
			}
			return []validator.Validator{
				&DuplicateTagValidator{Name: tag.Name, Uid: uid},
			}
		},
	}
	view.Run()
}

var getTagListHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := blueprint.ListView{
		Context:      context,
		Pagination:   &blueprint.DefaultPagination{},
		QueryBuilder: &service.TagQueryBuilder{},
		FilterMapping: []blueprint.FilterMapping{
			{
				Lookup: "video",
				Method: "InVideoIds",
				Many:   true,
			},
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &BaseTagTemplate{}
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &BaseListContainer{}
		},
		OnApplyQuery: func(v *blueprint.ListView) {
			v.QueryBuilder.(*service.TagQueryBuilder).Uid = context.Param["uid"].(string)
			context.BindingInput(v.QueryBuilder)
		},
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
	}
	view.Run()
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

package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/blueprint"
	"github.com/allentom/haruka/serializer"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
	"strconv"
)

var readVideoList haruka.RequestHandler = func(context *haruka.Context) {
	view := blueprint.ListView{
		Context:      context,
		Pagination:   &blueprint.DefaultPagination{},
		QueryBuilder: &service.VideoQueryBuilder{},
		FilterMapping: []blueprint.FilterMapping{
			{
				Lookup: "tag",
				Method: "InTagIds",
				Many:   true,
			},
			{
				Lookup: "library",
				Method: "InLibraryIds",
				Many:   true,
			},
		},
		OnApplyQuery: func(v *blueprint.ListView) {
			context.BindingInput(v.QueryBuilder)
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &BaseVideoTemplate{}
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
	}
	view.Run()
}

var getVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	videoPermissionValidator := VideoAccessibleValidator{}
	context.BindingInput(&videoPermissionValidator)
	if err = validator.RunValidators(&videoPermissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	video, err := service.GetVideoById(uint(id), context.Param["uid"].(string), "Files")
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseVideoTemplate{}
	template.Assign(video)
	context.JSON(template)
}

var deleteVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permission := VideoAccessibleValidator{}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.DeleteVideoById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type MoveVideoRequest struct {
	Path    string `json:"path"`
	Library uint   `json:"library"`
}

var moveVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody MoveVideoRequest
	err = context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	movePermissionChecker := MoveLibraryPermissionValidator{
		SourceVideoId: uint(id),
		LibraryId:     requestBody.Library,
		Uid:           context.Param["uid"].(string),
	}
	if err = validator.RunValidators(&movePermissionChecker); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	video, err := service.MoveVideoById(uint(id), requestBody.Library, requestBody.Path)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseVideoTemplate{}
	template.Assign(video)
	context.JSON(template)
}

type VideoTranscodeRequest struct {
	Format string `json:"format"`
	Codec  string `json:"codec"`
}

var transcodeHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody VideoTranscodeRequest
	err = context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permission := VideoAccessibleValidator{}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.NewVideoTranscodeTask(uint(id), requestBody.Format, requestBody.Codec)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

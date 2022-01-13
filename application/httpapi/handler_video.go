package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/blueprint"
	"github.com/allentom/haruka/serializer"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
	"strconv"
	"time"
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
	video, err := service.GetVideoById(uint(id), "Files", "Infos")
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseVideoTemplate{}
	template.Assign(video)
	// get subject
	if video.SubjectId > 0 && config.Instance.YouLibraryConfig.Enable {
		response, err := service.GetSubjectById(video.SubjectId)
		if err == nil {
			template.Subject = &response.Data
		}
	}
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

type AddVideoMetaRequestBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var addVideoMetaHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	body := AddVideoMetaRequestBody{}
	err = context.ParseJson(&body)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}

	permission := VideoAccessibleValidator{
		Id: uint(rawId),
	}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	meta, err := service.AddVideoInfoItem(uint(rawId), body.Key, body.Value)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(meta)
}

func removeVideoMetaHandler(context *haruka.Context) {
	rawId, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.RemoveInfoItem(uint(rawId))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type UpdateVideoRequestBody struct {
	Release  string `json:"release"`
	EntityId uint   `json:"entityId"`
}

var updateVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permission := VideoAccessibleValidator{
		Id: uint(rawId),
	}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	body := UpdateVideoRequestBody{}
	err = context.ParseJson(&body)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	updateData := make(map[string]interface{})
	if body.Release != "" {
		releaseTime, err := time.Parse("2006-01-02", body.Release)
		if err != nil {
			AbortError(context, err, http.StatusBadRequest)
			return
		}
		updateData["release"] = releaseTime
	}
	if body.EntityId != 0 {
		updateData["entity_id"] = body.EntityId
	}

	err = service.UpdateVideo(uint(rawId), updateData)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}

	newVideo, err := service.GetVideoById(uint(rawId))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseVideoTemplate{}
	template.Assign(newVideo)
	context.JSON(haruka.JSON{
		"success": true,
		"data":    template,
	})
}

var getMetaListHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.InfoQueryBuilder{}
	err := context.BindingInput(&queryBuilder)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	infos, count, err := queryBuilder.Query()
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success":  true,
		"count":    count,
		"page":     queryBuilder.Page,
		"pageSize": queryBuilder.PageSize,
		"result":   serializer.SerializeMultipleTemplate(infos, &BaseVideoMetaTemplate{}, map[string]interface{}{}),
	})

}

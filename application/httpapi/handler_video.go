package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
	"strconv"
	"time"
)

func checkVideoAccessibleAndRaiseError(context *haruka.Context) bool {
	permission := VideoAccessibleValidator{}
	err := context.BindingInput(&permission)
	if err := validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return false
	}
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return false
	}
	return true
}

var readVideoList haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.VideoQueryBuilder{}
	if err := context.BindingInput(&queryBuilder); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	count, models, err := queryBuilder.Query()
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data := make([]BaseVideoTemplate, 0)
	for _, video := range models {
		template := BaseVideoTemplate{}
		template.Assign(video)
		data = append(data, template)
	}
	SendSuccessResponse(context, haruka.JSON{
		"count":    count,
		"page":     queryBuilder.Page,
		"pageSize": queryBuilder.PageSize,
		"result":   data,
	})
}

var getVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	if !checkVideoAccessibleAndRaiseError(context) {
		return
	}
	id, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	video, err := service.GetVideoById(uint(id), "Files", "Infos", "Files.Subtitles")
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseVideoTemplate{}
	template.Assign(video)
	// get subject
	if video.SubjectId != nil && config.Instance.YouLibraryConfig.Enable {
		response, err := service.GetSubjectById(*video.SubjectId)
		if err == nil {
			template.Subject = &response.Data
		}
	}
	SendSuccessResponse(context, template)
}

var deleteVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	if !checkVideoAccessibleAndRaiseError(context) {
		return
	}
	id, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.DeleteVideoById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, nil)
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
	SendSuccessResponse(context, template)
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
	if !checkVideoAccessibleAndRaiseError(context) {
		return
	}
	err = service.NewVideoTranscodeTask(uint(id), requestBody.Format, requestBody.Codec)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, nil)
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

	if !checkVideoAccessibleAndRaiseError(context) {
		return
	}
	_, err = service.AddVideoInfoItem(uint(rawId), body.Key, body.Value)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, nil)
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
	SendSuccessResponse(context, nil)
}

type UpdateVideoRequestBody struct {
	Release  string `json:"release"`
	EntityId uint   `json:"entityId"`
	Episode  string `json:"episode"`
	Order    uint   `json:"order"`
	Name     string `json:"name"`
}

var updateVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	if !checkVideoAccessibleAndRaiseError(context) {
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
	if len(body.Episode) > 0 {
		updateData["episode"] = body.Episode
	}
	if body.Order > 0 {
		updateData["order"] = body.Order
	}
	if len(body.Name) > 0 {
		updateData["name"] = body.Name
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
	SendSuccessResponse(context, template)
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
	SendSuccessResponse(context, haruka.JSON{
		"count":    count,
		"page":     queryBuilder.Page,
		"pageSize": queryBuilder.PageSize,
		"result":   serializer.SerializeMultipleTemplate(infos, &BaseVideoMetaTemplate{}, map[string]interface{}{}),
	})
}

var refreshVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.RefreshVideo(uint(rawId))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, nil)
}

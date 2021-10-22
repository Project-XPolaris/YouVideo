package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/youplus"
	"net/http"
	"strconv"
)

type CreateLibraryRequest struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Private   bool   `json:"private"`
	VideoType string `json:"videoType"`
}

var createLibraryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateLibraryRequest
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	uid := context.Param["uid"].(string)
	if !requestBody.Private {
		uid = service.PublicUid
	}
	libraryPath := requestBody.Path
	if config.Instance.YouPlusPath {
		realPath, err := youplus.DefaultClient.GetRealPath(requestBody.Path, context.Param["token"].(string))
		if err != nil {
			AbortError(context, err, http.StatusBadRequest)
			return
		}
		libraryPath = realPath
	}
	if err = validator.RunValidators(
		&DuplicateLibraryPathValidator{Path: libraryPath},
		&LibraryPathAccessibleValidator{Path: libraryPath},
	); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}

	library, err := service.CreateLibrary(libraryPath, requestBody.Name, uid, requestBody.VideoType)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseLibraryTemplate{}
	template.Assign(library)
	context.JSON(template)
}

var readLibraryList haruka.RequestHandler = func(context *haruka.Context) {
	page := context.Param["page"].(int)
	pageSize := context.Param["pageSize"].(int)
	queryBuilder := service.LibraryQueryOption{
		Page:     page,
		PageSize: pageSize,
	}
	err := context.BindingInput(&queryBuilder)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	count, libraryList, err := service.GetLibraryList(queryBuilder)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data := make([]BaseLibraryTemplate, 0)
	for _, library := range libraryList {
		template := BaseLibraryTemplate{}
		template.Assign(&library)
		data = append(data, template)
	}
	context.JSON(haruka.JSON{
		"count":    count,
		"page":     page,
		"pageSize": pageSize,
		"result":   data,
	})
}

var scanLibrary haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	username := context.Param["username"].(string)
	permission := LibraryAccessibleValidator{}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	task, err := service.CreateSyncLibraryTask(service.CreateScanTaskOption{
		LibraryId: uint(id),
		Uid:       context.Param["uid"].(string),
		OnFileComplete: func(task *service.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventSyncTaskFileComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnFileError: func(task *service.Task, err error) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventSyncTaskFileError,
				"error": err,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnError: func(task *service.Task, err error) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventSyncTaskError,
				"error": err,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnComplete: func(task *service.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventSyncTaskComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
	})
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"task":    NewTaskTemplate(task),
	})
}
var readMetaTask haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	uid := context.Param["uid"].(string)
	username := context.Param["username"].(string)
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permission := LibraryAccessibleValidator{}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	_, err = service.CreateGenerateVideoMetaTask(service.CreateGenerateMetaOption{
		LibraryId: uint(id),
		Uid:       uid,
		OnFileComplete: func(task *service.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventMetaTaskFileComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnVideoComplete: func(task *service.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventMetaTaskVideoComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnComplete: func(task *service.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventMetaTaskComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnFileError: func(task *service.Task, err error) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventMetaTaskFileError,
				"error": err,
				"task":  NewTaskTemplate(task),
			}, username)
		},
	})
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
var deleteLibrary haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permission := LibraryAccessibleValidator{}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.RemoveLibraryById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

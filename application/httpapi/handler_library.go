package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/commons"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/service/task"
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
		realPath, err := plugin.DefaultYouPlusPlugin.Client.GetRealPath(requestBody.Path, context.Param["token"].(string))
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
		"success":  true,
		"count":    count,
		"page":     page,
		"pageSize": pageSize,
		"result":   data,
	})
}

type ScanLibraryRequestBody struct {
	MatchSubject bool `json:"matchSubject"`
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
	err = context.BindingInput(&permission)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody ScanLibraryRequestBody
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = context.ParseJson(&requestBody)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, &commons.APIError{
			Err:  err,
			Code: commons.CodeValidatorError,
			Desc: err.Error(),
		}, http.StatusBadRequest)
		return
	}
	task, err := task.CreateSyncLibraryTask(task.CreateScanTaskOption{
		LibraryId:    uint(id),
		Uid:          context.Param["uid"].(string),
		MatchSubject: requestBody.MatchSubject,
		OnFileComplete: func(task *task.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventSyncTaskFileComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnFileError: func(task *task.Task, err error) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventSyncTaskFileError,
				"error": err,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnError: func(task *task.Task, err error) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventSyncTaskError,
				"error": err,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnComplete: func(task *task.Task) {
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
var newRemoveLibraryTask haruka.RequestHandler = func(context *haruka.Context) {
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
		AbortError(context, &commons.APIError{
			Err:  err,
			Code: commons.CodeValidatorError,
			Desc: err.Error(),
		}, http.StatusBadRequest)
		return
	}
	task, err := task.CreateRemoveLibraryTask(task.RemoveLibraryOption{
		LibraryId: uint(id),
		Uid:       context.Param["uid"].(string),
		OnError: func(task *task.Task, err error) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventRemoveTaskError,
				"error": err,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnComplete: func(task *task.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventRemoveTaskComplete,
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
	_, err = task.CreateGenerateVideoMetaTask(task.CreateGenerateMetaOption{
		LibraryId: uint(id),
		Uid:       uid,
		OnFileComplete: func(task *task.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventMetaTaskFileComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnVideoComplete: func(task *task.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventMetaTaskVideoComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnComplete: func(task *task.Task) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventMetaTaskComplete,
				"task":  NewTaskTemplate(task),
			}, username)
		},
		OnFileError: func(task *task.Task, err error) {
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

package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
)

type FileObjectInput struct {
	Id uint `hsource:"path" hname:"id"`
}

var playVideo haruka.RequestHandler = func(context *haruka.Context) {
	var fileObjectInput FileObjectInput
	err := context.BindingInput(&fileObjectInput)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	filePermissionValidator := FilePermissionValidator{}
	err = context.BindingInput(&filePermissionValidator)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	if err = validator.RunValidators(&filePermissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	file, err := service.GetFileById(fileObjectInput.Id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	service.CreateHistory(file.VideoId, filePermissionValidator.Uid)
	http.ServeFile(context.Writer, context.Request, file.Path)
}

var videoSubtitle haruka.RequestHandler = func(context *haruka.Context) {
	var fileObjectInput FileObjectInput
	err := context.BindingInput(&fileObjectInput)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	filePermissionValidator := FilePermissionValidator{}
	err = context.BindingInput(&filePermissionValidator)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	if err = validator.RunValidators(&filePermissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	file, err := service.GetFileById(fileObjectInput.Id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	http.ServeFile(context.Writer, context.Request, file.Subtitles)
}
var removeFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var fileObjectInput FileObjectInput
	err := context.BindingInput(&fileObjectInput)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permissionValidator := FilePermissionValidator{}
	err = context.BindingInput(&permissionValidator)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	if err = validator.RunValidators(&permissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.DeleteFile(fileObjectInput.Id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type RenameFileRequestBody struct {
	Name string `json:"name"`
}

var renameFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var fileObjectInput FileObjectInput
	err := context.BindingInput(&fileObjectInput)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	permissionValidator := FilePermissionValidator{}
	err = context.BindingInput(&permissionValidator)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	if err = validator.RunValidators(&permissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody RenameFileRequestBody
	err = context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.RenameFile(fileObjectInput.Id, requestBody.Name)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

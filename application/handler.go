package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
	"strconv"
)

type CreateLibraryRequest struct {
	Path string `json:"path"`
}

var createLibraryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateLibraryRequest
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	library, err := service.CreateLibrary(requestBody.Path)
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
	count, libraryList, err := service.GetLibraryList(service.LibraryQueryOption{
		Page:     page,
		PageSize: pageSize,
	})
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
	err = service.ScanLibraryById(uint(id))
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
	err = service.RemoveLibraryById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var readVideoList haruka.RequestHandler = func(context *haruka.Context) {
	page := context.Param["page"].(int)
	pageSize := context.Param["pageSize"].(int)
	count, videoList, err := service.GetVideoList(service.VideoQueryOption{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data := make([]BaseVideoTemplate, 0)
	for _, video := range videoList {
		template := BaseVideoTemplate{}
		template.Assign(&video)
		data = append(data, template)
	}
	context.JSON(haruka.JSON{
		"count":    count,
		"page":     page,
		"pageSize": pageSize,
		"result":   data,
	})
}

var playVideo haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	video, err := service.GetVideoById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	http.ServeFile(context.Writer, context.Request, video.Path)
}

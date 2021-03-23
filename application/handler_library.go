package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
	"strconv"
)

type CreateLibraryRequest struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Private bool   `json:"private"`
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
	if err = validator.RunValidators(
		&DuplicateLibraryPathValidator{Path: requestBody.Path},
		&LibraryPathAccessibleValidator{Path: requestBody.Path},
	); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	library, err := service.CreateLibrary(requestBody.Path, requestBody.Name, uid)
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
	permission := LibraryAccessibleValidator{}
	context.BindingInput(&permission)
	if err = validator.RunValidators(&permission); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.ScanLibraryById(uint(id), context.Param["uid"].(string))
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

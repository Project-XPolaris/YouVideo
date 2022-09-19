package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
)

var matchVideoInformationHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetQueryInt("id")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.MatchVideoInformationById(id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type SearchMovieInformationInput struct {
	Query  string `hsource:"query" hname:"query"`
	Source string `hsource:"query" hname:"source"`
}

var searchMovieInformationHandler haruka.RequestHandler = func(context *haruka.Context) {
	var input SearchMovieInformationInput
	err := context.BindingInput(&input)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	source := service.GetInfoSource(input.Source)
	result, err := source.SearchMovieList(input.Query)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success":  true,
		"count":    len(result),
		"page":     1,
		"pageSize": len(result),
		"result":   NewSearchMoveInformationTemplates(result, input.Source),
	})
}

type SearchTvInformationInput struct {
	Query  string `hsource:"query" hname:"query"`
	Source string `hsource:"query" hname:"source"`
}

var searchTvInformationHandler haruka.RequestHandler = func(context *haruka.Context) {
	var input SearchTvInformationInput
	err := context.BindingInput(&input)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	source := service.GetInfoSource(input.Source)
	result, err := source.SearchTvList(input.Query)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success":  true,
		"count":    len(result),
		"page":     1,
		"pageSize": len(result),
		"result":   NewSearchTvInformationTemplates(result, input.Source),
	})
}

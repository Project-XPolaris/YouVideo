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

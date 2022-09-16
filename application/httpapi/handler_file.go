package httpapi

import (
	"bytes"
	context2 "context"
	"errors"
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
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
	service.CreateHistory(file.VideoId, context.Param["uid"].(string))
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

var fileCoverHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetPathParameterAsInt("id")
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
	file, err := service.GetFileById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	storageKey := filepath.Join(config.Instance.CoversStore, file.Cover)
	storage := plugin.GetDefaultStorage()
	buf, err := storage.Get(context2.Background(), plugin.GetDefaultBucket(), storageKey)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data, _ := ioutil.ReadAll(buf)
	http.ServeContent(context.Writer, context.Request, file.Cover, time.Now(), bytes.NewReader(data))
}

var playLinkHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.GetPathParameterAsString("id")
	sourcesType := context.GetPathParameterAsString("type")
	token := context.GetPathParameterAsString("token")
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	rawAuth, err := module.Auth.ParseToken(token)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	user := rawAuth.(*database.User)
	filePermissionValidator := FilePermissionValidator{
		Id:  uint(id),
		Uid: user.Uid,
	}
	if err = validator.RunValidators(&filePermissionValidator); err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	file, err := service.GetFileById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	switch sourcesType {
	case "video":
		service.CreateHistory(file.VideoId, user.Uid)
		file, err := os.OpenFile(file.Path, os.O_RDONLY, 0666)
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		http.ServeContent(context.Writer, context.Request, file.Name(), time.Now(), file)
		//http.ServeFile(context.Writer, context.Request, file.Path)
	case "subs":
		http.ServeFile(context.Writer, context.Request, file.Subtitles)
	case "cover":
		storageKey := filepath.Join(config.Instance.CoversStore, file.Cover)
		storage := plugin.GetDefaultStorage()
		buf, err := storage.Get(context2.Background(), plugin.GetDefaultBucket(), storageKey)
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		data, _ := ioutil.ReadAll(buf)
		http.ServeContent(context.Writer, context.Request, file.Cover, time.Now(), bytes.NewReader(data))
	case "cc":
		ccs, err := service.GetCloseCaption(file)
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		context.JSON(haruka.JSON{
			"success": true,
			"subs":    NewCCTemplates(ccs),
		})
		break
	default:
		AbortError(context, errors.New("invalid sources type"), http.StatusBadRequest)
	}

}

package application

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/youtrans"
	"net/http"
	"os"
)

var readDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	rootPath := context.GetQueryString("path")
	if len(rootPath) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		rootPath = homeDir
	}
	infos, err := service.ReadDirectory(rootPath)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data := make([]BaseFileItemTemplate, 0)
	for _, info := range infos {
		template := BaseFileItemTemplate{}
		template.Assign(info, rootPath)
		data = append(data, template)
	}
	context.JSON(haruka.JSON{
		"path":  rootPath,
		"sep":   string(os.PathSeparator),
		"files": data,
	})
}

var readTaskListHandler haruka.RequestHandler = func(context *haruka.Context) {
	tasks := service.GetTaskList()
	data := make([]BaseTaskTemplate, 0)
	for _, task := range tasks {
		template := BaseTaskTemplate{}
		template.Assign(task)
		data = append(data, template)
	}
	if config.Instance.EnableTranscode {
		transTaskResponse, err := youtrans.DefaultYouTransClient.GetTaskList()
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		for _, transTask := range transTaskResponse.List {
			template := BaseTaskTemplate{}
			template.AssignWithTrans(transTask)
			data = append(data, template)
		}
	}
	context.JSON(haruka.JSON{
		"count":  len(tasks),
		"result": data,
	})
}

var getCodecsHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.CodecsQueryBuilder{}
	err := context.BindingInput(&queryBuilder)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	codecs, err := queryBuilder.Query()
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	result := serializer.SerializeMultipleTemplate(codecs, &BaseCodecTemplate{}, nil)
	context.JSON(haruka.JSON{
		"codecs": result,
	})

}

var getFormatsHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.FormatsQueryBuilder{}
	err := context.BindingInput(&queryBuilder)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	formats, err := queryBuilder.Query()
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	result := serializer.SerializeMultipleTemplate(formats, &BaseFormatTemplate{}, nil)
	context.JSON(haruka.JSON{
		"formats": result,
	})

}

var transCompleteCallback haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody youtrans.TaskResponse
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.CompleteTrans(requestBody)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var serviceInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"name":        "YouVideo serivce",
		"authEnable":  config.Instance.EnableAuth,
		"authUrl":     fmt.Sprintf("%s/%s", config.Instance.AuthURL, "user/auth"),
		"transEnable": config.Instance.EnableTranscode,
	})
}

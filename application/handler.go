package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/blueprint"
	"github.com/allentom/haruka/serializer"
	"github.com/allentom/haruka/validator"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/youtrans"
	"net/http"
	"os"
	"strconv"
)

type CreateLibraryRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

var createLibraryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateLibraryRequest
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	library, err := service.CreateLibrary(requestBody.Path, requestBody.Name)
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
	view := blueprint.ListView{
		Context:      context,
		Pagination:   &blueprint.DefaultPagination{},
		QueryBuilder: &service.VideoQueryBuilder{},
		FilterMapping: []blueprint.FilterMapping{
			{
				Lookup: "tag",
				Method: "InTagIds",
				Many:   true,
			},
			{
				Lookup: "library",
				Method: "InLibraryIds",
				Many:   true,
			},
		},
		OnApplyQuery: func(v *blueprint.ListView) {
			context.BindingInput(v.QueryBuilder)
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &BaseVideoTemplate{}
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
	}
	view.Run()
}

var getVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	video, err := service.GetVideoById(uint(id), "Files")
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseVideoTemplate{}
	template.Assign(video)
	context.JSON(template)
}
var playVideo haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	file, err := service.GetFileById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	http.ServeFile(context.Writer, context.Request, file.Path)
}

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

var deleteVideoHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.DeleteVideoById(uint(id))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
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
	video, err := service.MoveVideoById(uint(id), requestBody.Library, requestBody.Path)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseVideoTemplate{}
	template.Assign(video)
	context.JSON(template)
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
	err = service.NewVideoTranscodeTask(uint(id), requestBody.Format, requestBody.Codec)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type CreateTagRequestBody struct {
	Name string `json:"name"`
}

var createTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := blueprint.CreateModelView{
		Context: context,
		CreateModel: func() interface{} {
			return &database.Tag{}
		},
		ResponseTemplate: &BaseTagTemplate{},
		RequestBody:      &CreateTagRequestBody{},
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},

		OnCreate: func(view *blueprint.CreateModelView, model interface{}) (interface{}, error) {
			tag := model.(*database.Tag)
			err := tag.Save()
			return tag, err
		},
		GetValidators: func(v *blueprint.CreateModelView) []validator.Validator {
			tag := v.RequestBody.(*CreateTagRequestBody)
			return []validator.Validator{
				&DuplicateTagValidator{Name: tag.Name},
			}
		},
	}
	view.Run()
}

var getTagListHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := blueprint.ListView{
		Context:      context,
		Pagination:   &blueprint.DefaultPagination{},
		QueryBuilder: &service.TagQueryBuilder{},
		FilterMapping: []blueprint.FilterMapping{
			{
				Lookup: "video",
				Method: "InVideoIds",
				Many:   true,
			},
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &BaseTagTemplate{}
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &BaseListContainer{}
		},
		OnApplyQuery: func(v *blueprint.ListView) {
			context.BindingInput(v.QueryBuilder)
		},
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
	}
	view.Run()
}
var removeTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := blueprint.DeleteModelView{
		Context: context,
		Lookup:  "id",
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
		Model: &database.Tag{},
	}
	view.Run()
}

var updateTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := blueprint.UpdateModelView{
		Context: context,
		Lookup:  "id",
		OnError: func(err error) {
			AbortError(context, err, http.StatusInternalServerError)
			return
		},
		Model:    &database.Tag{},
		Template: &BaseTagTemplate{},
	}
	view.Run()
}

type TagVideoBatchRequestBody struct {
	Ids []uint
}

var addVideoToTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody TagVideoBatchRequestBody
	err = context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.AddVideosToTag(uint(id), requestBody.Ids...)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var removeVideoFromTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawId := context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	var requestBody TagVideoBatchRequestBody
	err = context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.RemoveVideosFromTag(uint(id), requestBody.Ids...)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
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

type TagVideosRequestBody struct {
	Name []string `json:"name"`
	Ids  []uint   `json:"ids"`
}

var tagVideosBatchHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody TagVideosRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	err = service.AddOrCreateTagFromVideo(requestBody.Name, requestBody.Ids...)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

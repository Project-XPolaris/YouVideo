package task

import (
	"errors"
	"github.com/allentom/harukap/module/task"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/util"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CreateVideoTaskOption struct {
	Uid          string
	FilePath     string
	libraryId    uint
	CreateOption *service.CreateVideoFileOptions
	ParentTaskId string
}
type CreateVideoTaskOutput struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
}
type CreateVideoTask struct {
	*task.BaseTask
	option     *CreateVideoTaskOption
	TaskOutput *CreateVideoTaskOutput
	PathList   []string
	video      *database.Video
}

func (t *CreateVideoTask) Stop() error {
	return nil
}

func (t *CreateVideoTask) Start() error {
	videoLogger := plugin.DefaultYouLogPlugin.Logger.NewScope("create video")
	option := t.option.CreateOption
	if option == nil {
		option = &service.CreateVideoFileOptions{}
	}

	file, err := service.GetFileByPath(t.option.FilePath)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return t.AbortError(err)
	}
	// create file if not exist
	if file == nil {
		file = &database.File{}
	}

	// check if file is updated
	md5Task := CreateMD5Task(MD5Option{
		Uid:          t.option.Uid,
		FilePath:     t.option.FilePath,
		ParentTaskId: t.Id,
	})
	t.SubTaskList = append(t.SubTaskList, md5Task)
	err = task.RunTask(md5Task)
	if err != nil {
		return t.AbortError(err)
	}
	fileCheckSum := md5Task.CheckSum
	isUpdate := fileCheckSum != file.Checksum
	file.Checksum = fileCheckSum

	// check if video file exist
	file.Path = t.option.FilePath

	// save folder index
	baseDir := filepath.Dir(t.option.FilePath)
	var folder database.Folder
	err = database.Instance.FirstOrCreate(&folder, database.Folder{Path: baseDir, LibraryId: t.option.libraryId}).Error
	if err != nil {
		return t.AbortError(err)
	}

	// get or create video
	fileExt := filepath.Ext(t.option.FilePath)
	videoName := strings.TrimSuffix(filepath.Base(t.option.FilePath), fileExt)
	var video database.Video
	err = database.Instance.Model(&database.Video{}).Where("name = ?", videoName).Where("base_dir = ?", baseDir).First(&video).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return t.AbortError(err)
	}
	// create if not found
	if video.ID == 0 {
		video = database.Video{
			Name:      videoName,
			LibraryId: t.option.libraryId,
			BaseDir:   baseDir,
			FolderID:  &folder.ID,
		}
		videoLogger.WithFields(youlog.Fields{
			"name": videoName,
		}).Info("video not exist,try to create")
		err = database.Instance.Create(&video).Error
		if err != nil {
			return t.AbortError(err)
		}
	}
	// match subject
	//if t.option.CreateOption.matchSubject && isUpdate {
	//	DefaultVideoInformationMatchService.In <- NewVideoInformationMatchInput(&video)
	//}
	if *video.FolderID != folder.ID {
		video.FolderID = &folder.ID
		err = database.Instance.Save(video).Error
		if err != nil {
			return t.AbortError(err)
		}
	}
	file.VideoId = video.ID
	err = database.Instance.Save(file).Error
	if err != nil {
		return t.AbortError(err)
	}

	generateCoverTask := NewGenerateCoverTask(&GenerateCoverTaskOption{
		file:         file,
		Uid:          t.option.Uid,
		ParentTaskId: t.ParentTaskId,
	})
	t.SubTaskList = append(t.SubTaskList, generateCoverTask)
	err = task.RunTask(generateCoverTask)
	if err != nil {
		return t.AbortError(err)
	}
	err = database.Instance.Save(file).Error
	if err != nil {
		return t.AbortError(err)
	}

	// read subtitles
	if isUpdate {
		items, err := os.ReadDir(baseDir)
		if err != nil {
			return t.AbortError(err)
		}
		for _, item := range items {
			if util.IsSubtitlesFile(item.Name()) && strings.HasPrefix(item.Name(), videoName+".") {
				_, err = database.ReadOrCreateSubtitles(filepath.Join(baseDir, item.Name()), file.ID)
				if err != nil {
					return t.AbortError(err)
				}
			}
		}
	}

	if isUpdate {
		analyzeFileMetaTask := NewAnalyzeFileMetaTask(&AnalyzeFileMetaTaskOption{
			Uid:          t.option.Uid,
			path:         t.option.FilePath,
			ParentTaskId: t.Id,
		})
		t.SubTaskList = append(t.SubTaskList, analyzeFileMetaTask)
		err = task.RunTask(analyzeFileMetaTask)
		if err != nil {
			videoLogger.Error(err)
		}
		if analyzeFileMetaTask.MetaData != nil {
			meta := analyzeFileMetaTask.MetaData
			file.Duration = meta.Format.DurationSeconds
			size, err := strconv.ParseInt(meta.Format.Size, 10, 64)
			if err != nil {
				videoLogger.Error(err)
			} else {
				file.Size = size
			}
			bitrate, err := strconv.ParseInt(meta.Format.BitRate, 10, 64)
			if err != nil {
				videoLogger.Error(err)
			} else {
				file.Bitrate = bitrate
			}
			// parse stream
			for _, stream := range meta.Streams {
				if stream.CodecType == "video" && len(file.MainVideoCodec) == 0 {
					file.MainVideoCodec = stream.CodecName
					continue
				}
				if stream.CodecType == "audio" && len(file.MainAudioCodec) == 0 {
					file.MainAudioCodec = stream.CodecName
				}
			}

			err = database.Instance.Save(file).Error
			if err != nil {
				videoLogger.Error()
			}
		}
	}

	// analyze nsfw content
	if (isUpdate || option.ForceNSFWCheck) && plugin.DefaultNSFWCheckPlugin.Enable && option.EnableNSFWCheck {
		nsfwTask := NewNSFWCheckTask(&NSFWCheckTaskOption{
			Uid:          t.option.Uid,
			path:         t.option.FilePath,
			ParentTaskId: t.Id,
		})
		t.SubTaskList = append(t.SubTaskList, nsfwTask)
		err = task.RunTask(nsfwTask)
		if err != nil {
			videoLogger.Error(err)
		}
		if nsfwTask.Result != nil {
			video.Hentai = nsfwTask.Result.Hentai
			video.Sexy = nsfwTask.Result.Sexy
			video.Porn = nsfwTask.Result.Porn
			err = database.Instance.Save(file).Error
			if err != nil {
				videoLogger.Error(err)
			}
		}
	}
	if err != nil {
		return t.AbortError(err)
	}
	t.video = &video
	return nil
}

func (t *CreateVideoTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func NewCreateVideoTask(option *CreateVideoTaskOption) *CreateVideoTask {
	t := &CreateVideoTask{
		BaseTask: task.NewBaseTask(TaskTypeNameMapping[TaskCreateVideo], option.Uid, task.GetStatusText(nil, task.StatusRunning)),
		PathList: []string{},
		TaskOutput: &CreateVideoTaskOutput{
			Filename: filepath.Base(option.FilePath),
			Path:     option.FilePath,
		},
		option: option,
	}
	t.ParentTaskId = option.ParentTaskId
	return t
}

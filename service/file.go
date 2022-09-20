package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
	"github.com/projectxpolaris/youvideo/youtrans"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func RemoveFileById(id uint) error {
	var file database.File
	err := database.Instance.Model(&database.File{}).Where("id = ?", id).First(&file).Error
	if err != nil {
		return err
	}
	if len(file.Cover) > 0 && file.AutoGenCover {
		var coverRefCount int64
		err = database.Instance.Model(&database.File{}).Where("cover = ?", file.Cover).Count(&coverRefCount).Error
		if err != nil {
			return err
		}
		if coverRefCount == 1 {
			if util.CheckFileExist(file.Cover) {
				err = os.Remove(file.Cover)
				if err != nil {
					return err
				}
			}
		}
	}
	err = database.Instance.Unscoped().Where("id = ?", id).Delete(&database.File{}).Error
	return err
}

func GetFileById(id uint) (*database.File, error) {
	var file database.File
	err := database.Instance.Preload("Subtitles").Find(&file, id).Error
	return &file, err
}

func GetFileByPath(path string) (*database.File, error) {
	file := database.File{}
	err := database.Instance.Where("path = ?", path).Preload("Subtitles").First(&file).Error
	return &file, err
}

func NewFileTranscodeTask(id uint, format string, codec string) error {
	file, err := GetFileById(id)
	if err != nil {
		return err
	}
	oid := xid.New().String()
	outputFilename := filepath.Base(file.Path)
	ext := filepath.Ext(file.Path)
	outputFilename = strings.TrimSuffix(outputFilename, ext)
	outputFilename += fmt.Sprintf("_%s_%s.%s", oid, codec, format)
	output := filepath.Join(filepath.Dir(file.Path), outputFilename)
	response, err := youtrans.DefaultYouTransClient.CreateNewTask(&youtrans.CreateTaskRequestBody{
		Input:  file.Path,
		Output: output,
		Format: format,
		Codec:  codec,
	})
	if response != nil {
		logrus.Info(response.Id)
	}
	return err
}
func NewVideoFile(path string) (file database.File) {
	meta, metaerr := GetVideoFileMeta(path)
	if metaerr == nil {
		file.Duration = meta.Format.DurationSeconds
		size, err := strconv.ParseInt(meta.Format.Size, 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Size = size
		}
		bitrate, err := strconv.ParseInt(meta.Format.BitRate, 10, 64)
		if err != nil {
			VideoLogger.Error(err)
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
	}
	file.Path = path
	return
}
func CompleteTrans(tranTask youtrans.TaskResponse) error {
	original, err := GetFileByPath(tranTask.Input)
	if err != nil {
		return err
	}
	if original == nil {
		return errors.New("original trans file not found")
	}
	newFile := NewVideoFile(tranTask.Output)
	newFile.VideoId = original.VideoId
	err = database.Instance.Create(&newFile).Error
	if err != nil {
		return err
	}
	// generate cover
	if len(original.Cover) > 0 {
		coverFileName := fmt.Sprintf("%s%s", xid.New(), filepath.Ext(original.Cover))
		storage := plugin.GetDefaultStorage()
		err = storage.Copy(context.Background(),
			plugin.GetDefaultBucket(),
			path.Join(config.Instance.CoversStore, original.Cover),
			plugin.GetDefaultBucket(),
			path.Join(config.Instance.CoversStore, coverFileName),
		)
		if err != nil {
			return err
		}
		newFile.Cover = filepath.Base(coverFileName)
	}

	if len(original.Subtitles) > 0 {
		subName := strings.ReplaceAll(filepath.Base(tranTask.Output), filepath.Ext(tranTask.Output), "")
		for _, subs := range original.Subtitles {
			subFilename := util.ChangeFileNameWithoutExt(subs.Path, subName)
			subPath := filepath.Join(filepath.Dir(subs.Path), subFilename)
			err = util.CopyFile(subs.Path, subPath)
			if err != nil {
				return err
			}
			_, err := database.ReadOrCreateSubtitles(subPath, newFile.ID)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func DeleteFile(id uint) error {
	var file database.File
	err := database.Instance.First(&file, id).Error
	if err != nil {
		return err
	}
	var video database.Video
	err = database.Instance.Preload("Files").First(&video, file.VideoId).Error
	if err != nil {
		return err
	}
	err = RemoveFileById(id)
	if err != nil {
		return err
	}
	if len(video.Files) == 1 {
		err = database.Instance.Unscoped().Delete(&database.Video{}, video.ID).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func RenameFile(id uint, name string) error {
	var file database.File
	err := database.Instance.First(&file, id).Preload("Video").Error
	if err != nil {
		return err
	}
	exist, err := afero.Exists(AppFs, file.Path)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	fileDir := filepath.Dir(file.Path)
	fileExt := filepath.Ext(file.Path)
	newFilePath := filepath.Join(fileDir, fmt.Sprintf("%s%s", name, fileExt))
	err = AppFs.Rename(file.Path, newFilePath)
	if err != nil {
		return err
	}
	if len(file.Subtitles) > 0 {
		for _, sub := range file.Subtitles {
			subFileExt := filepath.Ext(sub.Path)
			newSubPath := filepath.Join(fileDir, fmt.Sprintf("%s%s", name, subFileExt))
			err = AppFs.Rename(sub.Path, newSubPath)
			if err != nil {
				return err
			}
			_, err = database.ReadOrCreateSubtitles(newSubPath, file.ID)
			if err != nil {
				return err
			}
		}

	}
	file.Path = newFilePath
	err = database.Instance.Save(&file).Error
	if err != nil {
		return err
	}
	return nil
}

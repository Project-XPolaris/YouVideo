package service

import (
	"errors"
	"fmt"
	"github.com/allentom/haruka/gormh"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/util"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var VideoLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "Service.Video",
})

func CheckLibrary(libraryId uint) error {
	library, err := GetLibraryById(libraryId, "Videos.Files")
	if err != nil {
		return err
	}
	for _, video := range library.Videos {
		removeCount := 0
		for _, file := range video.Files {
			if !util.CheckFileExist(file.Path) {
				err = RemoveFileById(file.ID)
				if err != nil {
					return err
				}
				removeCount++
			}
		}
		if removeCount == len(video.Files) {
			err = DeleteVideoById(video.ID)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
func ScanVideo(library *database.Library) error {
	targetExtensions := []string{
		"mp4", "mkv",
	}
	err := afero.Walk(AppFs, library.Path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, extension := range targetExtensions {
			if strings.HasSuffix(info.Name(), extension) {
				err = CreateVideoFile(path, library.ID)
				if err != nil {
					logrus.Error(err)
				}
				break
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return err
}

type VideoQueryBuilder struct {
	gormh.DefaultPageFilter
	VideoTagIdFilter
	VideoLibraryIdFilter
	Orders   []string `hsource:"query" hname:"order"`
	GroupBy  []string `hsource:"query" hname:"group"`
	BaseDirs []string `hsource:"query" hname:"dir"`
	Search   string   `hsource:"query" hname:"search"`
	Uid      string   `hsource:"param" hname:"uid"`
}

func (v *VideoQueryBuilder) InTagIds(ids ...interface{}) {
	if v.VideoTagIdFilter.tagIds == nil {
		v.VideoTagIdFilter.tagIds = []interface{}{}
	}
	v.VideoTagIdFilter.tagIds = append(v.VideoTagIdFilter.tagIds, ids...)
}
func (v *VideoQueryBuilder) InLibraryIds(ids ...interface{}) {
	if v.VideoLibraryIdFilter.libraryIds == nil {
		v.VideoLibraryIdFilter.libraryIds = []interface{}{}
	}
	v.VideoLibraryIdFilter.libraryIds = append(v.VideoLibraryIdFilter.libraryIds, ids...)
}
func (v *VideoQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = gormh.ApplyFilters(v, query)
	for _, order := range v.Orders {
		query = query.Order(order)
	}
	for _, group := range v.GroupBy {
		query = query.Distinct(group)
	}
	if v.BaseDirs != nil && len(v.BaseDirs) > 0 {
		query = query.Where("base_dir IN ?", v.BaseDirs)
	}
	if len(v.Search) > 0 {
		query = query.Where("name like ?", "%"+v.Search+"%")
	}
	query = query.Joins("left join library_users on library_users.library_id = videos.library_id")
	query = query.Joins("left join users on library_users.user_id = users.id")
	if len(v.Uid) > 0 {
		query = query.Where("users.uid in ?", []string{v.Uid, PublicUid})
	} else {
		query = query.Where("users.uid in ?", []string{PublicUid})
	}
	models := make([]*database.Video, 0)
	var count int64
	err := query.Model(&database.Video{}).
		Preload("Files").
		Limit(v.GetLimit()).
		Offset(v.GetOffset()).
		Find(&models).
		Offset(-1).
		Count(&count).
		Error
	return count, models, err
}

type VideoTagIdFilter struct {
	tagIds []interface{}
}

func (f VideoTagIdFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.tagIds != nil && len(f.tagIds) > 0 {
		return db.Joins("left join video_tags on video_tags.video_id = videos.id").Where("video_tags.tag_id In ?", f.tagIds)
	}
	return db
}

type VideoLibraryIdFilter struct {
	libraryIds []interface{}
}

func (f VideoLibraryIdFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.libraryIds != nil && len(f.libraryIds) > 0 {
		return db.Where("videos.library_id In ?", f.libraryIds)
	}
	return db
}
func CreateVideoFile(path string, libraryId uint) error {
	// check if video file exist
	var video database.Video
	file, err := GetFileByPath(path)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	fileExt := filepath.Ext(path)
	videoName := strings.TrimSuffix(filepath.Base(path), fileExt)
	baseDir := filepath.Dir(path)
	VideoLogger.WithField("filename", videoName).Info("file hit")

	if file == nil {
		file = &database.File{}
	}

	file.Path = path

	// get meta data
	meta, metaerr := GetVideoFileMeta(path)
	if metaerr == nil {
		duration, err := strconv.ParseFloat(meta.GetFormat().GetDuration(), 10)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Duration = duration
		}

		size, err := strconv.ParseInt(meta.GetFormat().GetSize(), 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Size = size
		}

		bitrate, err := strconv.ParseInt(meta.GetFormat().GetBitRate(), 10, 64)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Bitrate = bitrate
		}

		// parse stream
		for _, stream := range meta.GetStreams() {
			if stream.GetCodecType() == "video" && len(file.MainVideoCodec) == 0 {
				file.MainVideoCodec = stream.GetCodecName()
				continue
			}
			if stream.GetCodecType() == "audio" && len(file.MainAudioCodec) == 0 {
				file.MainAudioCodec = stream.GetCodecName()
			}
		}
	}

	// check cover
	needGenerate := true
	targetCoverFilePaths := []string{
		"cover.jpg",
		"cover.png",
		"cover.jpeg",
		"cover.JPEG",
		"cover.PNG",
		fmt.Sprintf("%s.jpg", videoName),
		fmt.Sprintf("%s.png", videoName),
		fmt.Sprintf("%s.jpeg", videoName),
		fmt.Sprintf("%s.JPEG", videoName),
		fmt.Sprintf("%s.PNG", videoName),
	}
	for _, targetCoverFilePath := range targetCoverFilePaths {
		coverSourcePath := filepath.Join(baseDir, targetCoverFilePath)
		if util.CheckFileExist(coverSourcePath) {
			VideoLogger.Info(fmt.Sprintf("use exist cover  = %s", coverSourcePath))
			coverFileName := fmt.Sprintf("%s%s", xid.New(), filepath.Ext(coverSourcePath))
			savePath, err := filepath.Abs(filepath.Join(config.Instance.CoversStore, coverFileName))
			if err != nil {
				VideoLogger.Error(err)
				break
			}

			err = util.CopyFile(coverSourcePath, savePath)
			if err != nil {
				VideoLogger.Error(err)
				break
			}

			os.Remove(filepath.Join(config.Instance.CoversStore, file.Cover))
			file.Cover = coverFileName
			needGenerate = false
			break
		}
	}

	if needGenerate && len(file.Cover) == 0 {
		coverPath, err := GenerateVideoCover(path)
		if err != nil {
			VideoLogger.Error(err)
		} else {
			file.Cover = filepath.Base(coverPath)
		}
	}

	// check file subtitles file
	subtitlesFiles := []string{
		fmt.Sprintf("%s.srt", videoName),
		fmt.Sprintf("%s.ass", videoName),
		fmt.Sprintf("%s.ssa", videoName),
	}
	subtitlesFilePath := ""
	for _, subtitlesFile := range subtitlesFiles {
		subtitlesSourcePath := filepath.Join(baseDir, subtitlesFile)
		if util.CheckFileExist(subtitlesSourcePath) {
			subtitlesFilePath = subtitlesSourcePath
			break
		}
	}
	file.Subtitles = subtitlesFilePath
	if file.VideoId == 0 {
		err = database.Instance.Model(&database.Video{}).Where("name = ?", videoName).Where("base_dir = ?", baseDir).First(&video).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// create if not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			video = database.Video{
				Name:      videoName,
				LibraryId: libraryId,
				BaseDir:   baseDir,
			}
			VideoLogger.WithFields(logrus.Fields{
				"name": videoName,
			}).Warn("video not exist,try to create")
			err = database.Instance.Create(&video).Error
			if err != nil {
				return err
			}
		}
		file.VideoId = video.ID
	}
	err = database.Instance.Save(file).Error
	return err
}

func DeleteVideoById(id uint) error {
	var video database.Video
	// remove files
	err := database.Instance.Preload("Files").First(&video, id).Error
	if err != nil {
		return err
	}
	for _, file := range video.Files {
		err = RemoveFileById(file.ID)
		if err != nil {
			return err
		}
	}
	// clean tag rel
	err = database.Instance.Model(&video).Association("Tags").Clear()
	if err != nil {
		return err
	}
	err = database.Instance.
		Model(&database.Video{}).
		Unscoped().
		Where("id = ?", id).
		Delete(&database.Video{}).
		Error
	if err != nil {
		return err
	}
	return nil
}

type VideoQueryOption struct {
	Page      int
	PageSize  int
	WithFiles bool
}

func GetVideoList(option VideoQueryOption) (int64, []database.Video, error) {
	var result []database.Video
	var count int64
	queryBuilder := database.Instance.Model(&database.Video{})
	if option.WithFiles {
		queryBuilder.Preload("Files")
	}
	err := queryBuilder.Limit(option.PageSize).Count(&count).Offset((option.Page - 1) * option.PageSize).Find(&result).Error
	return count, result, err
}

func GetVideoById(id uint, uid string, rel ...string) (*database.Video, error) {
	var video database.Video
	query := database.Instance
	for _, relStr := range rel {
		query = query.Preload(relStr)
	}
	query = query.Joins("left join library_users on library_users.library_id = videos.library_id").
		Joins("left join users on users.id = library_users.user_id").
		Where("users.uid in ?", []string{PublicUid, uid})
	err := query.First(&video, id).Error
	return &video, err
}

func MoveVideoById(id uint, targetLibraryId uint, targetPath string) (*database.Video, error) {
	var video database.Video
	err := database.Instance.Model(&database.Video{}).Where("id = ?", id).Preload("Files").First(&video).Error
	if err != nil {
		return nil, err
	}
	// load source library
	sourceLibrary, err := GetLibraryById(video.LibraryId)
	if err != nil {
		return nil, err
	}

	// load target library
	targetLibrary, err := GetLibraryById(targetLibraryId)
	if err != nil {
		return nil, err
	}

	// move files
	for _, file := range video.Files {

		if len(targetPath) == 0 {
			targetPath, err = util.GetMovePath(file.Path, sourceLibrary.Path, targetLibrary.Path)
			if err != nil {
				return nil, err
			}
		}
		err = AppFs.MkdirAll(filepath.Dir(targetPath), os.ModePerm)
		if err != nil {
			return nil, err
		}
		// move file
		err = AppFs.Rename(file.Path, targetPath)
		if err != nil {
			return nil, err
		}
		file.Path = targetPath
		// move subtitles
		if len(file.Subtitles) > 0 {
			err = AppFs.Rename(file.Subtitles, filepath.Join(filepath.Dir(targetPath), filepath.Base(file.Subtitles)))
		}
		database.Instance.Save(&file)
		video.BaseDir = filepath.Dir(targetPath)
	}

	video.LibraryId = targetLibraryId

	return &video, database.Instance.Save(&video).Error
}

func NewVideoTranscodeTask(id uint, uid string, format string, codec string) error {
	video, err := GetVideoById(id, uid, "Files")
	if err != nil {
		return err
	}
	if len(video.Files) > 0 {
		return NewFileTranscodeTask(video.Files[0].ID, format, codec)
	}
	return nil
}

func CheckVideoAccessible(id uint, uid string) bool {
	var videoCount int64
	database.Instance.
		Model(&database.Video{}).
		Joins("left join library_users on library_users.library_id = videos.library_id").
		Joins("left join users on library_users.user_id = users.id").
		Where("videos.id = ?", id).
		Where("users.uid in ?", []string{PublicUid, uid}).Count(&videoCount)
	return videoCount > 0
}

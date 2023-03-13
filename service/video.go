package service

import (
	"errors"
	"fmt"
	"github.com/allentom/haruka/gormh"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var VideoLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "Service.Video",
})

func RemoveNotExistVideo(libraryId uint) error {
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
func ScanVideo(library *database.Library, excludeDir []string) ([]string, error) {
	targetExtensions := []string{
		"mp4", "mkv", "avi", "rmvb", "flv", "wmv", "mov", "3gp", "m4v", "mpg", "mpeg", "mpe", "mpv", "m2v", "m4v", "m4p", "m4b", "m4r", "m4v", "m4a", "m4p", "m4b", "m4r", "m4v", "m4a", "m4p", "m4b", "m4r", "m4v", "m4a",
	}
	target := make([]string, 0)
	err := filepath.Walk(library.Path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			isExclude := false
			for _, dir := range excludeDir {
				if info.Name() == dir {
					isExclude = true
					break
				}
			}
			if isExclude {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		if info.Size() == 0 {
			return nil
		}
		for _, extension := range targetExtensions {

			if strings.HasSuffix(info.Name(), extension) {
				target = append(target, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return target, err
}

type VideoQueryBuilder struct {
	Library          []int      `hsource:"query" hname:"library"`
	Tag              []int      `hsource:"query" hname:"tag"`
	Page             int        `hsource:"param" hname:"page"`
	PageSize         int        `hsource:"param" hname:"pageSize"`
	Orders           []string   `hsource:"query" hname:"order"`
	GroupBy          []string   `hsource:"query" hname:"group"`
	BaseDirs         []string   `hsource:"query" hname:"dir"`
	Search           string     `hsource:"query" hname:"search"`
	Uid              string     `hsource:"param" hname:"uid"`
	Random           string     `hsource:"query" hname:"random"`
	FolderId         uint       `hsource:"query" hname:"folder"`
	DirectoryVideoId uint       `hsource:"query" hname:"directoryVideo"`
	DirectoryView    uint       `hsource:"query" hname:"directoryView"`
	InfoId           uint       `hsource:"query" hname:"info"`
	ReleaseStart     *time.Time `hsource:"query" hname:"releaseStart" format:"2006-01-02"`
	ReleaseEnd       *time.Time `hsource:"query" hname:"releaseEnd" format:"2006-01-02"`
	MaxDuration      int        `hsource:"query" hname:"maxDuration"`
	MinDuration      int        `hsource:"query" hname:"minDuration"`
	NSFW             string     `hsource:"query" hname:"nsfw"`
}

func (v *VideoQueryBuilder) Query() (int64, []*database.Video, error) {
	query := database.Instance
	query = gormh.ApplyFilters(v, query)
	if len(v.Random) > 0 {
		if database.Instance.Dialector.Name() == "sqlite" {
			query = query.Order("random()")
		} else {
			query = query.Order("RAND()")
		}
	} else {
		for _, order := range v.Orders {
			query = query.Order(fmt.Sprintf("videos.%s", order))
		}
	}

	for _, group := range v.GroupBy {
		query = query.Group(group)
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
	if v.DirectoryVideoId > 0 {
		var video database.Video
		err := database.Instance.Where("id = ?", v.DirectoryVideoId).Find(&video).Error
		if err != nil {
			return 0, nil, err
		}
		query = query.Where("base_dir = ?", video.BaseDir)
	}
	if v.FolderId > 0 {
		query = query.Where("folder_id = ?", v.FolderId)
	}
	if v.InfoId > 0 {
		query = query.Joins("left join video_infos on video_infos.video_id = videos.id")
		query = query.Where("video_infos.video_meta_item_id = ?", v.InfoId)
	}
	if v.Tag != nil && len(v.Tag) > 0 {
		query = query.Joins("left join video_tags on video_tags.video_id = videos.id").
			Where("video_tags.tag_id In ?", v.Tag)
	}
	if v.Library != nil && len(v.Library) > 0 {
		query = query.Where("videos.library_id In ?", v.Library)
	}
	if v.ReleaseStart != nil {
		query = query.Where("release >= ?", v.ReleaseStart)
	}
	if v.ReleaseEnd != nil {
		query = query.Where("release < ?", v.ReleaseEnd)
	}
	if v.MaxDuration > 0 || v.MinDuration > 0 {
		query = query.Joins("left join files on files.video_id = videos.id")
		if v.MaxDuration > 0 {
			query = query.Where("files.duration <= ?", v.MaxDuration)
		}
		if v.MinDuration > 0 {
			query = query.Where("files.duration >= ?", v.MinDuration)
		}
	}

	switch v.NSFW {
	case "only":
		query = query.Where(
			database.Instance.Or("hentai = ?", true).
				Or("porn = ?", true).
				Or("sexy = ?", true),
		)
	case "exclude":
		query = query.Where("hentai = ?", false).
			Where("porn = ?", false).
			Where("sexy = ?", false)
	case "hentai":
		query = query.Where("hentai = ?", true)
	case "sexy":
		query = query.Where("sexy = ?", true)
	case "porn":
		query = query.Where("porn = ?", true)
	}

	models := make([]*database.Video, 0)
	var count int64
	err := query.Model(&database.Video{}).
		Preload("Files").
		Preload("Files.Subtitles").
		Preload("Infos").
		Limit(v.PageSize).
		Offset(v.PageSize * (v.Page - 1)).
		Find(&models).
		Offset(-1).
		Count(&count).
		Error
	return count, models, err
}

type CreateVideoFileOptions struct {
	ForceNSFWCheck bool `json:"ForceNSFWCheck"`
}

func CreateVideoFile(path string, libraryId uint, videoType string, matchSubject bool, option *CreateVideoFileOptions) (*database.Video, error) {
	if option == nil {
		option = &CreateVideoFileOptions{}
	}
	file, err := GetFileByPath(path)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// create file if not exist
	if file == nil {
		file = &database.File{}
	}

	// check if file is updated
	fileCheckSum, err := util.MD5Checksum(path)
	if err != nil {
		return nil, err
	}
	isUpdate := fileCheckSum != file.Checksum
	file.Checksum = fileCheckSum

	// check if video file exist
	file.Path = path

	// save folder index
	baseDir := filepath.Dir(path)
	var folder database.Folder
	err = database.Instance.FirstOrCreate(&folder, database.Folder{Path: baseDir, LibraryId: libraryId}).Error
	if err != nil {
		return nil, err
	}

	// get or create video
	fileExt := filepath.Ext(path)
	videoName := strings.TrimSuffix(filepath.Base(path), fileExt)
	var video database.Video
	err = database.Instance.Model(&database.Video{}).Where("name = ?", videoName).Where("base_dir = ?", baseDir).First(&video).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// create if not found
	if video.ID == 0 {
		video = database.Video{
			Name:      videoName,
			LibraryId: libraryId,
			BaseDir:   baseDir,
			Type:      videoType,
			FolderID:  &folder.ID,
		}
		VideoLogger.WithFields(logrus.Fields{
			"name": videoName,
		}).Info("video not exist,try to create")
		err = database.Instance.Create(&video).Error
		if err != nil {
			return nil, err
		}
	}
	// match subject
	if matchSubject && isUpdate {
		DefaultVideoInformationMatchService.In <- NewVideoInformationMatchInput(&video)
	}
	if *video.FolderID != folder.ID {
		video.FolderID = &folder.ID
		err = database.Instance.Save(video).Error
		if err != nil {
			return nil, err
		}
	}
	file.VideoId = video.ID
	err = database.Instance.Save(file).Error
	if err != nil {
		return nil, err
	}

	DefaultVideoCoverGenerator.In <- file
	err = database.Instance.Save(file).Error
	if err != nil {
		return nil, err
	}

	// read subtitles
	if isUpdate {
		items, err := os.ReadDir(baseDir)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if util.IsSubtitlesFile(item.Name()) && strings.HasPrefix(item.Name(), videoName+".") {
				_, err = database.ReadOrCreateSubtitles(filepath.Join(baseDir, item.Name()), file.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if isUpdate {
		meta, err := GetVideoFileMeta(file.Path)
		if err != nil {
			VideoLogger.Error(err)
		}
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

		err = database.Instance.Save(file).Error
		if err != nil {
			return nil, err
		}
	}

	// analyze nsfw content
	if (isUpdate || option.ForceNSFWCheck) && plugin.DefaultNSFWCheckPlugin.Enable {
		outChan := make(chan io.Reader)
		go func() {
			err := util.ExtractNShotFromVideoPipe(file.Path, config.Instance.NSFWCheckConfig.Slice, outChan)
			if err != nil {
				VideoLogger.WithFields(logrus.Fields{
					"error": err,
				}).Error("extract frame error")
			}
		}()

		for shot := range outChan {
			result, err := plugin.DefaultNSFWCheckPlugin.Client.Predict(shot)
			if err != nil {
				VideoLogger.WithFields(logrus.Fields{
					"error": err,
				}).Error("nsfw check error")
				continue
			}
			for _, predictions := range result {
				if predictions.Probability > 0.5 {
					switch predictions.Classname {
					case "Hentai":
						video.Hentai = true
						break
					case "Porn":
						video.Porn = true
						break
					case "Sexy":
						video.Sexy = true
						break
					}
				}
			}
			err = database.Instance.Save(video).Error
			if err != nil {
				return nil, err
			}
		}
	}

	return &video, err
}
func RefreshVideo(videoId uint) error {
	var video database.Video
	err := database.Instance.Preload("Files").Preload("Library").First(&video, videoId).Error
	if err != nil {
		return err
	}
	for _, file := range video.Files {
		_, err = CreateVideoFile(file.Path, video.LibraryId, video.Type, false, nil)
		if err != nil {
			return err
		}
	}
	return nil
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
	// clean info rel
	infoAss := database.Instance.Model(&video).Association("Infos")
	infos := make([]database.VideoMetaItem, 0)
	err = infoAss.Find(&infos)
	if err != nil {
		return err
	}
	err = database.Instance.Model(&video).Association("Infos").Clear()
	if err != nil {
		return err
	}
	infoIds := make([]uint, 0)
	for _, info := range infos {
		infoIds = append(infoIds, info.ID)
	}
	TryToRemoveEmptyRelInfo(infoIds...)
	err = database.Instance.
		Model(&database.Video{}).
		Unscoped().
		Where("id = ?", id).
		Delete(&database.Video{}).
		Error
	if err != nil {
		return err
	}
	// remove video info
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

func GetVideoById(id uint, rel ...string) (*database.Video, error) {
	var video database.Video
	query := database.Instance
	for _, relStr := range rel {
		query = query.Preload(relStr)
	}
	err := query.First(&video, id).Error
	return &video, err
}

func MoveVideoById(id uint, targetLibraryId uint, targetPath string) (*database.Video, error) {
	var video database.Video
	err := database.Instance.Model(&database.Video{}).
		Where("id = ?", id).
		Preload("Files").
		Preload("Files.Subtitles").
		First(&video).
		Error
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
			for _, sub := range file.Subtitles {
				newSubPath := filepath.Join(filepath.Dir(targetPath), filepath.Base(sub.Path))
				err = AppFs.Rename(sub.Path, newSubPath)
				if err != nil {
					return nil, err
				}
				_, err = database.ReadOrCreateSubtitles(newSubPath, file.ID)
				if err != nil {
					return nil, err
				}

			}
		}
		database.Instance.Save(&file)
		video.BaseDir = filepath.Dir(targetPath)
	}

	video.LibraryId = targetLibraryId

	return &video, database.Instance.Save(&video).Error
}

func NewVideoTranscodeTask(id uint, format string, codec string) error {
	video, err := GetVideoById(id, "Files")
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

type UpdateVideoData struct {
	Release time.Time
}

func UpdateVideo(id uint, updateData map[string]interface{}) error {
	query := database.Instance.Model(&database.Video{}).Where("id = ?", id)
	err := query.Updates(updateData).Error
	if err != nil {
		return err
	}
	return nil
}

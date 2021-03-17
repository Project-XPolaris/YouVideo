package application

import "github.com/projectxpolaris/youvideo/service"

type DuplicateTagValidator struct {
	Name string
	Uid  string `hsource:"param" hname:"uid"`
}

func (v *DuplicateTagValidator) Check() (string, bool) {
	tag, _ := service.GetTagByName(v.Name, v.Uid)
	if tag.ID > 0 {
		return "tag already exist!", false
	}
	return "", true
}

type TagOwnerPermission struct {
	Id  uint   `hsource:"path" hname:"id"`
	Uid string `hsource:"param" hname:"uid"`
}

func (v *TagOwnerPermission) Check() (string, bool) {
	tag, _ := service.GetTagByID(v.Id, v.Uid)
	if tag.ID == 0 {
		return "tag not exist or not accessible", false
	}
	return "", true
}

type VideoPermissionValidator struct {
	VideoId uint   `hsource:"path" hname:"id"`
	Uid     string `hsource:"param" hname:"uid"`
}

func (v *VideoPermissionValidator) Check() (string, bool) {
	video, err := service.GetVideoById(v.VideoId, v.Uid)
	if err != nil || video.ID == 0 {
		return "video not exist or not accessible", false
	}
	return "", true
}

type FilePermissionValidator struct {
	Id  uint   `hsource:"path" hname:"id"`
	Uid string `hsource:"param" hname:"uid"`
}

func (v *FilePermissionValidator) Check() (string, bool) {
	file, err := service.GetFileById(v.Id)
	if err != nil {
		return "file not exist or not accessible", false
	}
	video, err := service.GetVideoById(file.VideoId, v.Uid)
	if err != nil || video.ID == 0 {
		return "file not exist or not accessible", false
	}
	return "", true
}

type MoveLibraryPermissionValidator struct {
	SourceVideoId uint
	LibraryId     uint
	Uid           string
}

func (v *MoveLibraryPermissionValidator) Check() (string, bool) {
	canAccess := service.CheckLibraryUidOwner(v.LibraryId, v.Uid)
	if !canAccess {
		return "library not exist or not accessible", false
	}

	_, err := service.GetVideoById(v.SourceVideoId, v.Uid)
	if err != nil {
		return "video not exist or not accessible", false
	}
	return "", true
}

type LibraryAccessibleValidator struct {
	Id  uint   `hsource:"path" hname:"id"`
	Uid string `hsource:"param" hname:"uid"`
}

func (v *LibraryAccessibleValidator) Check() (string, bool) {
	canAccessible := service.CheckLibraryUidOwner(v.Id, v.Uid)
	if !canAccessible {
		return "library not exist or not accessible", false
	}
	return "", true
}

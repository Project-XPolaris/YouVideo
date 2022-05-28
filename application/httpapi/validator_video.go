package httpapi

import "github.com/projectxpolaris/youvideo/service"

type VideoAccessibleValidator struct {
	Id  uint   `hsource:"path" hname:"id"`
	Qid uint   `hsource:"query" hname:"id"`
	Uid string `hsource:"param" hname:"uid"`
}

func (v *VideoAccessibleValidator) Check() (string, bool) {
	videoId := v.Id
	if v.Qid > 0 {
		videoId = v.Qid
	}
	if service.CheckVideoAccessible(videoId, v.Uid) {
		return "", true
	}
	return "video not accessible", false
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
	if !service.CheckVideoAccessible(v.SourceVideoId, v.Uid) {
		return "video not exist or not accessible", false
	}
	return "", true
}

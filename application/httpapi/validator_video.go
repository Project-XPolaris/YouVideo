package httpapi

import "github.com/projectxpolaris/youvideo/service"

type VideoAccessibleValidator struct {
	Id  uint   `hsource:"path" hname:"id"`
	Uid string `hsource:"param" hname:"uid"`
}

func (v *VideoAccessibleValidator) Check() (string, bool) {
	if service.CheckVideoAccessible(v.Id, v.Uid) {
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

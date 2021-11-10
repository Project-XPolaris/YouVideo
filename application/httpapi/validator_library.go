package httpapi

import (
	"github.com/projectxpolaris/youvideo/service"
	"github.com/projectxpolaris/youvideo/util"
)

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

type DuplicateLibraryPathValidator struct {
	Path string
}

func (v *DuplicateLibraryPathValidator) Check() (string, bool) {
	isDuplicate := service.CheckLibraryPathExist(v.Path)
	if isDuplicate {
		return "library path already exist!", false
	}
	return "", true
}

type LibraryPathAccessibleValidator struct {
	Path string
}

func (v *LibraryPathAccessibleValidator) Check() (string, bool) {
	if util.CheckFileExist(v.Path) {
		return "", true
	}
	return "library path not accessible", false

}

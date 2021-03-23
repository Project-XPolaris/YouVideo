package application

import "github.com/projectxpolaris/youvideo/service"

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

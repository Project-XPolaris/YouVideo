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

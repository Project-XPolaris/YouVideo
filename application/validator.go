package application

import "github.com/projectxpolaris/youvideo/service"

type DuplicateTagValidator struct {
	Name string
}

func (v *DuplicateTagValidator) Check() (string, bool) {
	tag, _ := service.GetTagByName(v.Name)
	if tag.ID != 0 {
		return "tag already exist!", false
	}
	return "", true
}

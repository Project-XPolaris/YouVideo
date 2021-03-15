package service

import "github.com/projectxpolaris/youvideo/database"

const (
	PublicUid      = "-1"
	PublicUsername = "public"
)

func GetUserById(uid string) (*database.User, error) {
	user := database.User{
		Uid: uid,
	}
	err := database.Instance.Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

package plugin

import (
	"errors"
	"github.com/allentom/harukap/commons"
	"github.com/allentom/harukap/youplus"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
)

var DefaultYouPlusPlugin *youplus.Plugin

func CreateDefaultYouPlusPlugin() {
	DefaultYouPlusPlugin = &youplus.Plugin{}
	DefaultYouPlusPlugin.AuthFromToken = func(token string) (commons.AuthUser, error) {
		return GetUserByPlusAuthToken(token)
	}
	DefaultYouPlusPlugin.AuthUrl = "/oauth/youplus"
	module.Auth.Plugins = append(module.Auth.Plugins, DefaultYouPlusPlugin)
}
func GetUserByPlusAuthToken(accessToken string) (*database.User, error) {
	var oauthRecord database.Oauth
	response, err := DefaultYouPlusPlugin.Client.CheckAuth(accessToken)
	if err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, errors.New("invalid token")
	}
	err = database.Instance.Model(&database.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "YouPlusService").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}

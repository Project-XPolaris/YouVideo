package plugin

import (
	"github.com/allentom/harukap/commons"
	"github.com/allentom/harukap/youauth"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/module"
)

var DefaultYouAuthOauthPlugin *youauth.OauthPlugin

func CreateYouAuthPlugin() {
	DefaultYouAuthOauthPlugin = &youauth.OauthPlugin{}
	DefaultYouAuthOauthPlugin.AuthFromToken = func(token string) (commons.AuthUser, error) {
		return GetUserByYouAuthToken(token)
	}
	DefaultYouPlusPlugin.AuthUrl = "/oauth/youplus"
	module.Auth.Plugins = append(module.Auth.Plugins, DefaultYouAuthOauthPlugin)
}
func GetUserByYouAuthToken(accessToken string) (*database.User, error) {
	var oauthRecord database.Oauth
	err := database.Instance.Model(&database.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "youauth").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	_, err = DefaultYouAuthOauthPlugin.Client.GetCurrentUser(accessToken)
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}

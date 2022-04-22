package service

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/projectxpolaris/youvideo/database"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

const YouAuthProvider = "youauth"
const YouPlusAuthProvider = "YouPlusService"

func CreateRandomUser() (*database.User, error) {
	username := xid.New().String()
	// create new user
	user := &database.User{
		Uid:      username,
		Username: username,
	}
	err := database.Instance.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
func GenerateYouAuthToken(code string) (string, string, error) {
	tokens, err := plugin.DefaultYouAuthOauthPlugin.Client.GetAccessToken(code)
	if err != nil {
		return "", "", err
	}
	currentUserResponse, err := plugin.DefaultYouAuthOauthPlugin.Client.GetCurrentUser(tokens.AccessToken)
	if err != nil {
		return "", "", err
	}
	// check if user exists
	uid := fmt.Sprintf("%d", currentUserResponse.Id)
	historyOauth := make([]database.Oauth, 0)
	err = database.Instance.Where("uid = ?", uid).
		Where("provider = ?", YouAuthProvider).
		Preload("User").
		Find(&historyOauth).Error
	if err != nil {
		return "", "", err
	}
	var user *database.User
	if len(historyOauth) == 0 {
		user, err = CreateRandomUser()
		if err != nil {
			return "", "", err
		}
	} else {
		user = historyOauth[0].User
	}

	oauthRecord := database.Oauth{
		Uid:          fmt.Sprintf("%d", currentUserResponse.Id),
		UserId:       user.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Provider:     YouAuthProvider,
	}
	err = database.Instance.Create(&oauthRecord).Error
	if err != nil {
		return "", "", err
	}
	return tokens.AccessToken, currentUserResponse.Username, nil
}

func refreshToken(accessToken string) (string, error) {
	tokenRecord := database.Oauth{}
	err := database.Instance.Where("access_token = ?", accessToken).First(&tokenRecord).Error
	if err != nil {
		return "", err
	}
	token, err := plugin.DefaultYouAuthOauthPlugin.Client.RefreshAccessToken(tokenRecord.RefreshToken)
	if err != nil {
		return "", err
	}
	err = database.Instance.Delete(&tokenRecord).Error
	if err != nil {
		return "", err
	}
	newOauthRecord := database.Oauth{
		UserId:       tokenRecord.UserId,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	err = database.Instance.Create(&newOauthRecord).Error
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func LoginWithYouPlusAuth(username string, password string) (*database.Oauth, error) {
	resp, err := plugin.DefaultYouPlusPlugin.Client.FetchUserAuth(username, password)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New("invalid token")
	}
	var oauthRecord database.Oauth
	var user *database.User
	err = database.Instance.Model(&database.Oauth{}).Preload("User").Where("uid = ?", resp.Uid).
		Where("provider = ?", YouPlusAuthProvider).
		First(&oauthRecord).Error
	if err == gorm.ErrRecordNotFound {
		// create new user
		user, err = CreateRandomUser()
		if err != nil {
			return nil, err
		}
		oauthRecord = database.Oauth{
			Uid:          resp.Uid,
			Provider:     YouPlusAuthProvider,
			AccessToken:  resp.Token,
			RefreshToken: "",
			UserId:       user.ID,
		}
		err = database.Instance.Create(&oauthRecord).Error
		if err != nil {
			return nil, err
		}
		oauthRecord.User = user
	} else {
		if err != nil {
			return nil, err
		}
	}
	user = oauthRecord.User
	// save access token
	newOauth := database.Oauth{
		Uid:          resp.Uid,
		Provider:     YouPlusAuthProvider,
		AccessToken:  resp.Token,
		RefreshToken: "",
		UserId:       user.ID,
	}
	err = database.Instance.Create(&newOauth).Error
	if err != nil {
		return nil, err
	}
	return &oauthRecord, nil
}
func GetUserByAuthToken(accessToken string) (*database.User, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	mapClaims := token.Claims.(jwt.MapClaims)
	isu := mapClaims["iss"].(string)
	var oauthRecord database.Oauth
	switch isu {
	case YouAuthProvider:
		err = database.Instance.Model(&database.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
			Where("provider = ?", YouAuthProvider).
			Find(&oauthRecord).Error
		if err != nil {
			return nil, err
		}
		_, err = plugin.DefaultYouAuthOauthPlugin.Client.GetCurrentUser(accessToken)
		if err != nil {
			return nil, err
		}
		return oauthRecord.User, nil
	case YouPlusAuthProvider:
		response, err := plugin.DefaultYouPlusPlugin.Client.CheckAuth(accessToken)
		if err != nil {
			return nil, err
		}
		if !response.Success {
			return nil, errors.New("invalid token")
		}
		err = database.Instance.Model(&database.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
			Where("provider = ?", YouPlusAuthProvider).
			Find(&oauthRecord).Error
		if err != nil {
			return nil, err
		}
	}
	return nil, errors.New("invalid token")
}

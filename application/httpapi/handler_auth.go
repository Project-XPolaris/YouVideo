package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/project-xpolaris/youplustoolkit/youlink"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/service"
	"net/http"
)

type UserAuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var youPlusLoginHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody UserAuthRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	oauthToken, err := service.LoginWithYouPlusAuth(requestBody.Username, requestBody.Password)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"data": haruka.JSON{
			"accessToken": oauthToken.AccessToken,
			"username":    oauthToken.User.Uid,
		},
	})
}

var youPlusTokenHandler haruka.RequestHandler = func(context *haruka.Context) {
	// check token is valid
	token := context.GetQueryString("token")
	_, err := plugin.DefaultYouAuthOauthPlugin.Client.GetCurrentUser(token)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
var generateAccessCodeWithYouAuthHandler haruka.RequestHandler = func(context *haruka.Context) {
	code := context.GetQueryString("code")
	accessToken, username, err := service.GenerateYouAuthToken(code)
	if err != nil {
		youlink.AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"data": haruka.JSON{
			"accessToken": accessToken,
			"username":    username,
		},
	})
}

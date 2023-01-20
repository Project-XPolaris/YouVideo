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
	SendSuccessResponse(context, haruka.JSON{
		"accessToken": oauthToken.AccessToken,
		"username":    oauthToken.User.Uid,
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
	SendSuccessResponse(context, nil)
}
var generateAccessCodeWithYouAuthHandler haruka.RequestHandler = func(context *haruka.Context) {
	code := context.GetQueryString("code")
	accessToken, username, err := service.GenerateYouAuthToken(code)
	if err != nil {
		youlink.AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, haruka.JSON{
		"accessToken": accessToken,
		"username":    username,
	})
}

var generateAccessCodeWithYouAuthPasswordHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody UserAuthRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	accessToken, username, err := service.LoginWithYouAuth(requestBody.Username, requestBody.Password)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	SendSuccessResponse(context, haruka.JSON{
		"accessToken": accessToken,
		"username":    username,
	})
}

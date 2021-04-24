package auth

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/projectxpolaris/youvideo/config"
)

var DefaultAuthClient = AuthClient{}

type AuthClient struct {
}

type AuthResponse struct {
	Success  bool   `json:"success,omitempty"`
	Username string `json:"username,omitempty"`
	Uid      string `json:"uid,omitempty"`
}

func (c *AuthClient) CheckAuth(token string) (*AuthResponse, error) {
	var responseBody AuthResponse
	client := resty.New()
	_, err := client.R().SetQueryParam("token", token).SetResult(&responseBody).Get(fmt.Sprintf("%s/%s", config.Instance.AuthURL, "user/auth"))
	if err != nil {
		return nil, err
	}
	return &responseBody, nil
}

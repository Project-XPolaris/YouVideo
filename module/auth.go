package module

import (
	"github.com/allentom/haruka"
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/module/auth"
	"github.com/projectxpolaris/youvideo/config"
)

var Auth = &auth.AuthModule{
	Plugins: []harukap.AuthPlugin{},
}

func CreateAuthModule() {
	Auth.ConfigProvider = config.DefaultConfigProvider
	Auth.AuthMiddleware.RequestFilter = func(c *haruka.Context) bool {
		noAuthPattern := []string{
			"/oauth/youauth",
			"/oauth/youplus",
			"/info",
		}
		for _, pattern := range noAuthPattern {
			if c.Pattern == pattern {
				return true
			}
		}
		return false
	}
	Auth.InitModule()
}

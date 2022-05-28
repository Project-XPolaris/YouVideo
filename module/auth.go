package module

import (
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/module/auth"
	"github.com/projectxpolaris/youvideo/config"
)

var Auth = &auth.AuthModule{
	Plugins: []harukap.AuthPlugin{},
}

func CreateAuthModule() {
	Auth.ConfigProvider = config.DefaultConfigProvider
	Auth.NoAuthPath = []string{
		"/oauth/youauth",
		"/oauth/youplus",
		"/info",
	}
	Auth.InitModule()
}

package config

import "github.com/spf13/viper"

var Instance Config

type Config struct {
	Addr            string `json:"addr"`
	Application     string
	Instance        string
	CoversStore     string `json:"covers_store"`
	FfmpegBin       string `json:"ffmpeg_bin"`
	FfprobeBin      string `json:"ffprobe_bin"`
	YoutransURL     string `json:"youtrans_url"`
	EnableTranscode bool
	EnableAuth      bool `json:"enable_auth"`
	YouPlusPath     bool
	YouPlusUrl      string
	YouLogEnable    bool
	YouLogAddress   string
}

func ReadConfig() error {
	configer := viper.New()
	configer.AddConfigPath("./")
	configer.AddConfigPath("../")
	configer.SetConfigType("yaml")
	configer.SetConfigName("config")
	err := configer.ReadInConfig()
	if err != nil {
		return err
	}
	configer.SetDefault("addr", ":7600")
	configer.SetDefault("application", "YouVideo Core Service")
	configer.SetDefault("instance", "main")
	configer.SetDefault("cover_store", "./static/covers")
	configer.SetDefault("ffmpeg_bin", "ffmpeg")
	configer.SetDefault("ffprobe_bin", "ffprobe")
	configer.SetDefault("transcode.url", "")
	configer.SetDefault("transcode.enable", false)
	// auth
	configer.SetDefault("youplus.auth", false)
	configer.SetDefault("youplus.enablepath", false)
	configer.SetDefault("youplus.url", "")
	configer.SetDefault("youlog.enable", false)
	configer.SetDefault("youlog.rpc_addr", "")

	Instance = Config{
		Addr:            configer.GetString("addr"),
		Application:     configer.GetString("application"),
		Instance:        configer.GetString("instance"),
		CoversStore:     configer.GetString("cover_store"),
		FfmpegBin:       configer.GetString("ffmpeg_bin"),
		FfprobeBin:      configer.GetString("ffprobe_bin"),
		YoutransURL:     configer.GetString("transcode.url"),
		EnableTranscode: configer.GetBool("transcode.enable"),
		EnableAuth:      configer.GetBool("youplus.auth"),
		YouPlusPath:     configer.GetBool("youplus.enablepath"),
		YouPlusUrl:      configer.GetString("youplus.url"),
		YouLogEnable:    configer.GetBool("youlog.enable"),
		YouLogAddress:   configer.GetString("youlog.rpc_addr"),
	}
	return nil
}

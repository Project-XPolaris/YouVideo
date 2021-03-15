package config

import "github.com/spf13/viper"

var Instance Config

type Config struct {
	Addr            string `json:"addr"`
	CoversStore     string `json:"covers_store"`
	FfmpegBin       string `json:"ffmpeg_bin"`
	FfprobeBin      string `json:"ffprobe_bin"`
	YoutransURL     string `json:"youtrans_url"`
	EnableTranscode bool
	AuthURL         string `json:"auth_url"`
	EnableAuth      bool   `json:"enable_auth"`
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
	configer.SetDefault("cover_store", "./static/covers")
	configer.SetDefault("ffmpeg_bin", "ffmpeg")
	configer.SetDefault("ffprobe_bin", "ffprobe")
	configer.SetDefault("transcode.url", "")
	configer.SetDefault("transcode.enable", false)
	// auth
	configer.SetDefault("auth.enable", false)
	configer.SetDefault("auth.url", "")

	Instance = Config{
		Addr:            configer.GetString("addr"),
		CoversStore:     configer.GetString("cover_store"),
		FfmpegBin:       configer.GetString("ffmpeg_bin"),
		FfprobeBin:      configer.GetString("ffprobe_bin"),
		YoutransURL:     configer.GetString("transcode.url"),
		EnableTranscode: configer.GetBool("transcode.enable"),
		EnableAuth:      configer.GetBool("auth.enable"),
		AuthURL:         configer.GetString("auth.url"),
	}
	return nil
}

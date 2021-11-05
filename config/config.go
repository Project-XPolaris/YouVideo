package config

import (
	"github.com/allentom/harukap/config"
)

var DefaultConfigProvider *config.Provider

func InitConfigProvider(OnLoaded func(provider *config.Provider)) error {
	var err error
	DefaultConfigProvider, err = config.NewProvider(OnLoaded)
	return err
}

var Instance Config

type EntityConfig struct {
	Enable  bool
	Name    string
	Version int64
}
type Config struct {
	CoversStore     string `json:"covers_store"`
	FfmpegBin       string `json:"ffmpeg_bin"`
	FfprobeBin      string `json:"ffprobe_bin"`
	YoutransURL     string `json:"youtrans_url"`
	EnableTranscode bool
	EnableAuth      bool `json:"enable_auth"`
	YouPlusPath     bool
	YouPlusUrl      string
	YouPlusRPCAddr  string
	YouLogEnable    bool
	Entity          EntityConfig
	YouLogAddress   string
}

func ReadConfig(provider *config.Provider) {
	configer := provider.Manager
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
	configer.SetDefault("youplus.rpc", "")
	configer.SetDefault("youlog.enable", false)
	configer.SetDefault("youlog.rpc_addr", "")

	Instance = Config{
		CoversStore:     configer.GetString("cover_store"),
		FfmpegBin:       configer.GetString("ffmpeg_bin"),
		FfprobeBin:      configer.GetString("ffprobe_bin"),
		YoutransURL:     configer.GetString("transcode.url"),
		EnableTranscode: configer.GetBool("transcode.enable"),
		EnableAuth:      configer.GetBool("youplus.auth"),
		YouPlusPath:     configer.GetBool("youplus.enablepath"),
	}
}

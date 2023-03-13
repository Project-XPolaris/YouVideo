package config

import (
	"github.com/allentom/harukap/config"
	"os"
)

var DefaultConfigProvider *config.Provider

func InitConfigProvider() error {
	var err error
	customConfigPath := os.Getenv("YOUVIDEO_CONFIG_PATH")
	DefaultConfigProvider, err = config.NewProvider(func(provider *config.Provider) {
		ReadConfig(provider)
	}, customConfigPath)
	return err
}

var Instance Config

type EntityConfig struct {
	Enable  bool
	Name    string
	Version int64
}
type YouLibraryConfig struct {
	Enable bool
	Url    string
}
type TMdbConfig struct {
	Enable bool
	ApiKey string
	Proxy  string
}
type BangumiConfig struct {
	Enable bool
}
type NSFWCheckConfig struct {
	Slice int
}
type Config struct {
	CoversStore      string `json:"covers_store"`
	TempStore        string `json:"temp_store"`
	FfmpegBin        string `json:"ffmpeg_bin"`
	FfprobeBin       string `json:"ffprobe_bin"`
	YoutransURL      string `json:"youtrans_url"`
	EnableTranscode  bool
	YouPlusPath      bool
	YouLibraryConfig YouLibraryConfig
	TMdbConfig       TMdbConfig
	BangumiConfig    BangumiConfig
	SearchEngine     string          `json:"search_engine"`
	NSFWCheckConfig  NSFWCheckConfig `json:"nsfw_check_config"`
}

func ReadConfig(provider *config.Provider) {
	configer := provider.Manager
	configer.SetDefault("addr", ":7600")
	configer.SetDefault("application", "YouVideo Core Service")
	configer.SetDefault("instance", "main")
	configer.SetDefault("cover_store", "./static/covers")
	configer.SetDefault("temp_store", "./temp")
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
	configer.SetDefault("tmdb.enable", false)
	configer.SetDefault("bangumi.enable", false)
	configer.SetDefault("nsfwcheck.slice", 5)
	Instance = Config{
		CoversStore:     configer.GetString("cover_store"),
		FfmpegBin:       configer.GetString("ffmpeg_bin"),
		FfprobeBin:      configer.GetString("ffprobe_bin"),
		YoutransURL:     configer.GetString("transcode.url"),
		EnableTranscode: configer.GetBool("transcode.enable"),
		YouPlusPath:     configer.GetBool("youplus.enablepath"),
		TempStore:       configer.GetString("temp_store"),
		YouLibraryConfig: YouLibraryConfig{
			Enable: configer.GetBool("youlibrary.enable"),
			Url:    configer.GetString("youlibrary.url"),
		},
		TMdbConfig: TMdbConfig{
			Enable: configer.GetBool("tmdb.enable"),
			ApiKey: configer.GetString("tmdb.apikey"),
			Proxy:  configer.GetString("tmdb.proxy"),
		},
		BangumiConfig: BangumiConfig{
			Enable: configer.GetBool("bangumi.enable"),
		},
		SearchEngine: configer.GetString("search_engine"),
		NSFWCheckConfig: NSFWCheckConfig{
			Slice: configer.GetInt("nsfwcheck.slice"),
		},
	}
}

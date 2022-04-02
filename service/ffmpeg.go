package service

import (
	. "github.com/ahmetb/go-linq/v3"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)
import (
	"fmt"
	"github.com/allentom/transcoder"
	"github.com/allentom/transcoder/ffmpeg"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/rs/xid"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func NewTranscoder() transcoder.Transcoder {
	conf := &ffmpeg.Config{
		FfmpegBinPath:  config.Instance.FfmpegBin,
		FfprobeBinPath: config.Instance.FfprobeBin,
	}
	trans := ffmpeg.New(conf)
	return trans
}
func GetShotByFile(path string, output string) error {

	trans := NewTranscoder()
	trans.Input(path).Input(path)
	meta, err := trans.GetMetadata()
	if err != nil {
		return err
	}
	rawSeconds, err := strconv.ParseFloat(meta.GetFormat().GetDuration(), 10)
	if err != nil {
		return err
	}

	err = ffmpeg_go.
		Input(
			path,
			ffmpeg_go.KwArgs{"ss": fmt.Sprintf("%d", int(rawSeconds)/2)},
		).
		Output(
			output,
			ffmpeg_go.KwArgs{
				"vframes": "1",
				"vf":      "scale=320:-1",
				"q:v":     "2",
			},
		).Run()
	if err != nil {
		return err
	}
	return nil
}

func GenerateVideoCover(path string) (string, error) {
	err := os.MkdirAll(config.Instance.CoversStore, os.FileMode(0775))
	if err != nil {
		return "", err
	}
	outputPath, err := filepath.Abs(filepath.Join(config.Instance.CoversStore, fmt.Sprintf("%s.jpg", xid.New().String())))
	if err != nil {
		return "", err
	}
	err = GetShotByFile(path, outputPath)
	if err != nil {
		return "", err
	}
	return outputPath, err
}

func GetVideoFileMeta(path string) (transcoder.Metadata, error) {
	trans := NewTranscoder()
	trans.Input(path).Input(path)
	meta, err := trans.GetMetadata()
	if err != nil {
		return nil, err
	}
	return meta, nil
}

type CodecsQueryBuilder struct {
	Type   []string `hsource:"query" hname:"type"`
	Feat   []string `hsource:"query" hname:"feat"`
	Search string   `hsource:"query" hname:"search"`
}

func (b *CodecsQueryBuilder) Query() ([]ffmpeg.Codec, error) {
	codec, err := ffmpeg.ReadCodecList(&ffmpeg.Config{
		FfmpegBinPath:  config.Instance.FfmpegBin,
		FfprobeBinPath: config.Instance.FfprobeBin,
	})
	query := From(codec)
	if b.Type != nil && len(b.Type) > 0 {
		query = query.Where(func(i interface{}) bool {
			for _, targetType := range b.Type {
				c := i.(ffmpeg.Codec)
				if c.Flags.VideoCodec && targetType == "video" {
					return true
				}
				if c.Flags.AudioCodec && targetType == "audio" {
					return true
				}
				if c.Flags.SubtitleCodec && targetType == "subtitle" {
					return true
				}
			}
			return false
		})
	}
	if b.Feat != nil && len(b.Feat) > 0 {
		query = query.Where(func(i interface{}) bool {
			c := i.(ffmpeg.Codec)
			for _, feat := range b.Feat {
				if !c.Flags.Encoding && feat == "encode" {
					return false
				}
				if !c.Flags.Decoding && feat == "decode" {
					return false
				}
			}
			return true
		})
	}
	if len(b.Search) > 0 {
		query = query.Where(func(i interface{}) bool {
			return strings.Contains(i.(ffmpeg.Codec).Name, b.Search) || strings.Contains(i.(ffmpeg.Codec).Desc, b.Search)
		})
	}
	query.ToSlice(&codec)
	if err != nil {
		return nil, err
	}
	return codec, nil
}

type FormatsQueryBuilder struct {
	Search string `hsource:"query" hname:"search"`
}

func (b *FormatsQueryBuilder) Query() ([]ffmpeg.SupportFormat, error) {
	formats, err := ffmpeg.GetFormats(&ffmpeg.Config{
		FfmpegBinPath:  config.Instance.FfmpegBin,
		FfprobeBinPath: config.Instance.FfprobeBin,
	})
	query := From(formats)
	if len(b.Search) > 0 {
		query = query.Where(func(i interface{}) bool {
			return strings.Contains(i.(ffmpeg.SupportFormat).Name, b.Search) || strings.Contains(i.(ffmpeg.SupportFormat).Desc, b.Search)
		})
	}
	query.ToSlice(&formats)
	if err != nil {
		return nil, err
	}
	return formats, nil
}

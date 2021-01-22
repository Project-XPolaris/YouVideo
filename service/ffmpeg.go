package service

import (
	"fmt"
	"github.com/allentom/transcoder"
	"github.com/allentom/transcoder/ffmpeg"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/rs/xid"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func NewTranscoder() transcoder.Transcoder {
	conf := &ffmpeg.Config{
		FfmpegBinPath:  config.AppConfig.FfmpegBin,
		FfprobeBinPath: config.AppConfig.FfprobeBin,
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
	cmdStr := fmt.Sprintf("-ss %d -i %s -vframes 1 -q:v 2 %s", int(rawSeconds)/2, path, output)
	cmd := exec.Command(config.AppConfig.FfmpegBin, strings.Split(cmdStr, " ")...)
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func GenerateVideoCover(path string) (string, error) {
	err := os.MkdirAll("./static/covers", os.FileMode(0775))
	if err != nil {
		return "", err
	}
	outputPath, err := filepath.Abs(filepath.Join("./static/covers", fmt.Sprintf("%s.jpg", xid.New().String())))
	if err != nil {
		return "", err
	}
	err = GetShotByFile(path, outputPath)
	if err != nil {
		return "", err
	}
	return outputPath, err
}

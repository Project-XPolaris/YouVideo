package util

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"io"
	"os"
)

func ExtractNShotFromVideo(videoPath string, n int) ([]io.Reader, error) {
	probeOut, err := ffmpeg_go.Probe(videoPath)
	if err != nil {
		return nil, err
	}
	totalDuration := gjson.Get(probeOut, "format.duration").Float()
	out := make([]io.Reader, 0)
	for i := 0; i < n; i++ {
		buffer := bytes.NewBuffer(nil)
		cutTime := int(totalDuration) / n * i
		err = ffmpeg_go.
			Input(
				videoPath,
				ffmpeg_go.KwArgs{"ss": fmt.Sprintf("%d", int(cutTime))},
			).
			Output("pipe:", ffmpeg_go.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
			WithOutput(buffer, os.Stdout).
			Run()
		if err != nil {
			return nil, err
		}
		out = append(out, buffer)
	}
	return out, nil
}

func ExtractFrameFromVideoWithPerSecond(videoPath string, perSecond int) ([]io.Reader, error) {
	probeOut, err := ffmpeg_go.Probe(videoPath)
	if err != nil {
		return nil, err
	}
	totalDuration := gjson.Get(probeOut, "format.duration").Float()
	out := make([]io.Reader, 0)
	for i := 0; i < int(totalDuration); i += perSecond {
		buffer := bytes.NewBuffer(nil)
		err = ffmpeg_go.
			Input(
				videoPath,
				ffmpeg_go.KwArgs{"ss": fmt.Sprintf("%d", i)},
			).
			Output("pipe:", ffmpeg_go.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
			WithOutput(buffer, os.Stdout).
			Run()
		if err != nil {
			return nil, err
		}
		out = append(out, buffer)
	}
	return out, nil
}
func ExtractFrameFromVideoWithPerSecondPipe(videoPath string, perSecond int, out chan<- io.Reader) error {
	probeOut, err := ffmpeg_go.Probe(videoPath)
	if err != nil {
		return err
	}
	totalDuration := gjson.Get(probeOut, "format.duration").Float()
	for i := 0; i < int(totalDuration); i += perSecond {
		buffer := bytes.NewBuffer(nil)
		err = ffmpeg_go.
			Input(
				videoPath,
				ffmpeg_go.KwArgs{"ss": fmt.Sprintf("%d", i)},
			).
			Output("pipe:", ffmpeg_go.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
			WithOutput(buffer, os.Stdout).
			Run()
		if err != nil {
			return err
		}
		out <- buffer
	}
	close(out)
	return nil
}
func ExtractNShotFromVideoPipe(videoPath string, n int, out chan<- io.Reader) error {

	probeOut, err := ffmpeg_go.Probe(videoPath)
	if err != nil {
		return err
	}
	totalDuration := gjson.Get(probeOut, "format.duration").Float()
	for i := 0; i < n; i++ {
		buffer := bytes.NewBuffer(nil)
		fmt.Sprintf("extract %d in %d ", int(totalDuration)/n*i, int(totalDuration))
		cutTime := int(totalDuration) / n * i
		err = ffmpeg_go.
			Input(
				videoPath,
				ffmpeg_go.KwArgs{"ss": fmt.Sprintf("%d", int(cutTime))},
			).
			Output("pipe:", ffmpeg_go.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
			WithOutput(buffer, os.Stdout).
			Run()
		if err != nil {
			close(out)
			return err
		}
		out <- buffer
	}
	close(out)
	return nil
}

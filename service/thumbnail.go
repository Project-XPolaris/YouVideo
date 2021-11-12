package service

import (
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/rs/xid"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func GenerateThumbnail(coverFilePath string) (string, error) {
	file, err := os.Open(coverFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var img image.Image
	ext := strings.ToLower(filepath.Ext(coverFilePath))
	switch ext {
	case ".jpg":
		img, err = jpeg.Decode(file)
		if err != nil {
			return "", err
		}
	case ".png":
		img, err = png.Decode(file)
		if err != nil {
			return "", err
		}
	}
	if img == nil {
		return "", errors.New("unexpect image format")
	}

	rid := xid.New().String()
	outputName := fmt.Sprintf("%s%s", rid, ext)
	m := resize.Resize(320, 0, img, resize.Lanczos3)
	out, err := os.Create(filepath.Join(config.Instance.CoversStore, outputName))
	if err != nil {
		return "", err
	}
	defer out.Close()
	switch ext {
	case ".jpg":
		err = jpeg.Encode(out, m, nil)
		if err != nil {
			return "", err
		}
	case ".png":
		err = png.Encode(out, m)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("unknown output format")
	}
	return outputName, nil
}

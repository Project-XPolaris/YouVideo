package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/allentom/harukap/plugins/thumbnail"
	"github.com/nfnt/resize"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/rs/xid"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func GenerateThumbnail(coverFilePath string) (string, error) {
	rid := xid.New().String()
	ext := strings.ToLower(filepath.Ext(coverFilePath))
	outputName := fmt.Sprintf("%s%s", rid, ext)
	output := filepath.Join(config.Instance.CoversStore, outputName)
	switch config.Instance.ThumbnailType {
	case "thumbnailservice":
		buf, err := plugin.DefaultThumbnailPlugin.Client.GenerateAsRaw(coverFilePath, output, thumbnail.ThumbnailOption{
			MaxWidth: 320,
			Mode:     "width",
		})
		if err != nil {
			return "", err
		}
		storage := plugin.GetDefaultStorage()
		err = storage.Upload(context.Background(), buf, plugin.GetDefaultBucket(), output)
		if err != nil {
			return "", err
		}
		return outputName, nil
	default:
		err := GenerateThumbnailWithResize(coverFilePath, output)
		if err != nil {
			return "", err
		}
		return outputName, nil
	}
}
func GenerateThumbnailWithResize(coverFilePath string, output string) error {
	file, err := os.Open(coverFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	var img image.Image
	ext := strings.ToLower(filepath.Ext(coverFilePath))
	switch ext {
	case ".jpg":
		img, err = jpeg.Decode(file)
		if err != nil {
			return err
		}
	case ".png":
		img, err = png.Decode(file)
		if err != nil {
			return err
		}
	}
	if img == nil {
		return errors.New("unexpect image format")
	}

	m := resize.Resize(320, 0, img, resize.Lanczos3)
	buf := new(bytes.Buffer)
	if err != nil {
		return err
	}
	switch ext {
	case ".jpg":
		err = jpeg.Encode(buf, m, nil)
		if err != nil {
			return err
		}
	case ".png":
		err = png.Encode(buf, m)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown output format")
	}
	storage := plugin.GetDefaultStorage()
	err = storage.Upload(context.Background(), buf, plugin.GetDefaultBucket(), output)
	if err != nil {
		return err
	}
	return nil
}

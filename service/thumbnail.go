package service

import (
	"context"
	"github.com/allentom/harukap/plugins/thumbnail"
	"github.com/projectxpolaris/youvideo/config"
	"github.com/projectxpolaris/youvideo/plugin"
	"github.com/projectxpolaris/youvideo/util"
	"github.com/rs/xid"
	"os"
	"path/filepath"
)

func GenerateThumbnail(source string) (string, error) {
	id := xid.New().String()
	thumbnailFileName := util.ChangeFileNameWithoutExt(filepath.Base(source), id)
	output := filepath.Join(config.Instance.CoversStore, thumbnailFileName)
	file, err := os.Open(source)
	if err != nil {
		return "", err
	}
	out, err := plugin.DefaultThumbnailPlugin.Resize(context.Background(), file, thumbnail.ThumbnailOption{
		MaxWidth:  320,
		MaxHeight: 320,
	})
	if err != nil {
		return "", err
	}
	storage := plugin.GetDefaultStorage()
	err = storage.Upload(context.Background(), out, plugin.GetDefaultBucket(), output)
	if err != nil {
		return "", err
	}
	return thumbnailFileName, nil
}

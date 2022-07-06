package plugin

import (
	"context"
	"github.com/allentom/harukap/plugins/storage"
	"github.com/projectxpolaris/youvideo/config"
	"io"
)

var StorageEnginePlugin = &storage.Engine{}

func GetDefaultStorage() storage.FileSystem {
	defaultStorageName := config.DefaultConfigProvider.Manager.GetString("storage.default")
	return StorageEnginePlugin.GetStorage(defaultStorageName)
}

func GetDefaultBucket() string {
	defaultStorageName := config.DefaultConfigProvider.Manager.GetString("storage.defaultBucket")
	return defaultStorageName
}

func ReadFileBuffer(key string) (io.ReadCloser, error) {
	result, err := GetDefaultStorage().Get(context.Background(), GetDefaultBucket(), key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

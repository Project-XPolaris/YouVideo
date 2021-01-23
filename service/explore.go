package service

import (
	"github.com/spf13/afero"
	"os"
)

func ReadDirectory(root string) ([]os.FileInfo, error) {
	infos, err := afero.ReadDir(AppFs, root)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

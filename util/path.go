package util

import (
	"path/filepath"
)

func GetMovePath(path string, sourcePath string, targetPath string) (string, error) {
	result, err := filepath.Rel(sourcePath, path)
	if err != nil {
		return "", err
	}
	return filepath.Join(targetPath, result), nil
}

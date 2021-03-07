package util

import (
	"os"
	"path/filepath"
)

func GetMovePath(path string, sourcePath string, targetPath string) (string, error) {
	result, err := filepath.Rel(sourcePath, path)
	if err != nil {
		return "", err
	}
	return filepath.Join(targetPath, result), nil
}

func CheckFileExist(path string) bool {
	stat, _ := os.Stat(path)
	if stat != nil {
		return true
	}
	return false
}

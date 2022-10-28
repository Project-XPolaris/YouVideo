package util

import (
	"fmt"
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
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if stat != nil {
		return true
	}
	return false
}

func ChangeFileNameWithoutExt(filename string, newName string) string {
	baseName := filepath.Base(filename)
	ext := filepath.Ext(baseName)
	return fmt.Sprintf("%s%s", newName, ext)
}

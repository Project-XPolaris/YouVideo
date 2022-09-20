package util

import (
	"io"
	"os"
	"path/filepath"
)

func CopyFile(source string, dest string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func IsSubtitlesFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".srt" || ext == ".ass" || ext == ".ssa" || ext == ".vtt"
}

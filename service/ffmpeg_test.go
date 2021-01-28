package service

import (
	"fmt"
	"github.com/projectxpolaris/youvideo/config"
	"path/filepath"
	"testing"
)

func TestReadMeta(t *testing.T) {
	err := config.ReadConfig()
	if err != nil {
		t.Error(err)
	}
	meta, err := GetVideoFileMeta("C:\\Users\\Takay\\Desktop\\video_library\\video1.mp4")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(meta.GetFormat().GetDuration())
}

func TestRelatePath(t *testing.T) {

	path1 := "C:\\Users\\Takay\\Desktop\\video_library\\folder1\\folder2\\video1.mp4"
	libraryPath := "C:\\Users\\Takay\\Desktop\\video_library"
	targetLibraryPath := "C:\\Users\\Takay\\Desktop\\more_library\\video_library2"
	result, err := filepath.Rel(libraryPath, path1)
	if err != nil {
		t.Error(err)
	}
	targetPath := filepath.Join(targetLibraryPath, result)
	fmt.Println(filepath.Dir(targetPath))

}

func TestCheckDrive(t *testing.T) {
	path1 := "/home/aren/share"
	result := filepath.VolumeName(path1)
	fmt.Println(result)

}

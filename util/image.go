package util

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

func GetImageSize(input io.Reader) (int, int, error) {
	m, _, err := image.Decode(input)
	if err != nil {
		return 0, 0, err
	}
	g := m.Bounds()
	return g.Dx(), g.Dy(), nil
}

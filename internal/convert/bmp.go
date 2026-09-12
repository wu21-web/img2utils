package convert

import (
	"image"
	"io"

	"golang.org/x/image/bmp"
)

func encodeBMP(w io.Writer, img image.Image) error {
	return bmp.Encode(w, img)
}

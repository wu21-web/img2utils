package convert

import (
	"image"
	"image/jpeg"
	"io"
)

const jpegQuality = 90

func encodeJPEG(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: jpegQuality})
}

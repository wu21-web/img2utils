package convert

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Format string

const (
	JPEG Format = "jpeg"
	PNG  Format = "png"
	WebP Format = "webp"
)

var ErrUnsupportedFormat = errors.New("unsupported format")

func ParseFormat(ext string) (Format, error) {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return JPEG, nil
	case ".png":
		return PNG, nil
	case ".webp":
		return WebP, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedFormat, ext)
	}
}

func Convert(inputPath, outputPath string) error {
	format, err := ParseFormat(filepath.Ext(outputPath))
	if err != nil {
		return err
	}
	return convertTo(inputPath, outputPath, format)
}

func ToJPEG(inputPath, outputPath string) error {
	return convertTo(inputPath, outputPath, JPEG)
}

func ToPNG(inputPath, outputPath string) error {
	return convertTo(inputPath, outputPath, PNG)
}

func ToWebP(inputPath, outputPath string) error {
	return convertTo(inputPath, outputPath, WebP)
}

func convertTo(inputPath, outputPath string, format Format) error {
	src, err := decode(inputPath)
	if err != nil {
		return err
	}
	return encode(outputPath, src, format)
}

func decode(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return img, nil
}

func encode(path string, img image.Image, format Format) (err error) {
	var encodeFn func(io.Writer, image.Image) error
	switch format {
	case JPEG:
		encodeFn = encodeJPEG
	case PNG:
		encodeFn = encodePNG
	case WebP:
		return ErrWebPEncode
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedFormat, format)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := f.Close()
		if err != nil {
			os.Remove(path)
			return
		}
		err = closeErr
	}()

	return encodeFn(f, img)
}

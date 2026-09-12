package convert

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Format string

const DefaultMaxPixels = 25_000_000

const (
	JPEG Format = "jpeg"
	PNG  Format = "png"
	WebP Format = "webp"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrTooManyPixels     = errors.New("image has too many pixels")
)

func ParseFormat(name string) (Format, error) {
	switch strings.ToLower(strings.TrimPrefix(name, ".")) {
	case "jpg", "jpeg":
		return JPEG, nil
	case "png":
		return PNG, nil
	case "webp":
		return WebP, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedFormat, name)
	}
}

func (f Format) Ext() string {
	switch f {
	case JPEG:
		return ".jpg"
	case PNG:
		return ".png"
	case WebP:
		return ".webp"
	default:
		return ""
	}
}

func (f Format) Name() string {
	switch f {
	case JPEG:
		return "jpg"
	case PNG:
		return "png"
	case WebP:
		return "webp"
	default:
		return ""
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

func ConvertStream(r io.Reader, w io.Writer, format Format) error {
	return ConvertStreamLimit(r, w, format, DefaultMaxPixels)
}

func ConvertStreamLimit(r io.Reader, w io.Writer, format Format, maxPixels int) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read image: %w", err)
	}

	if maxPixels > 0 {
		config, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("decode image: %w", err)
		}
		pixels := int64(config.Width) * int64(config.Height)
		if pixels > int64(maxPixels) {
			return fmt.Errorf("%w: %dx%d exceeds %d", ErrTooManyPixels, config.Width, config.Height, maxPixels)
		}
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}
	return encodeTo(w, img, format)
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

	return encodeTo(f, img, format)
}

func encodeTo(w io.Writer, img image.Image, format Format) error {
	switch format {
	case JPEG:
		return encodeJPEG(w, img)
	case PNG:
		return encodePNG(w, img)
	case WebP:
		return ErrWebPEncode
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedFormat, format)
	}
}

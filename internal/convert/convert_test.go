package convert

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const webpFixture = "../testdata/sample.webp"

func TestParseFormat(t *testing.T) {
	tests := []struct {
		ext  string
		want Format
	}{
		{"jpg", JPEG},
		{".jpg", JPEG},
		{".JPG", JPEG},
		{".jpeg", JPEG},
		{"jpeg", JPEG},
		{".png", PNG},
		{".PNG", PNG},
		{"png", PNG},
		{".webp", WebP},
		{"webp", WebP},
	}

	for _, tt := range tests {
		got, err := ParseFormat(tt.ext)
		if err != nil {
			t.Fatalf("ParseFormat(%q) error: %v", tt.ext, err)
		}
		if got != tt.want {
			t.Errorf("ParseFormat(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}

	if _, err := ParseFormat(".bmp"); !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("ParseFormat(.bmp) error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestConvertStream(t *testing.T) {
	var input bytes.Buffer
	if err := png.Encode(&input, testImage()); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := ConvertStream(&input, &output, JPEG); err != nil {
		t.Fatalf("ConvertStream() error: %v", err)
	}

	img, format, err := image.Decode(&output)
	if err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" {
		t.Errorf("format = %q, want jpeg", format)
	}
	if got := img.Bounds().Dx(); got != 2 {
		t.Errorf("width = %d, want 2", got)
	}
}

func TestConvertStreamRejectsInvalidImage(t *testing.T) {
	var output bytes.Buffer
	if err := ConvertStream(strings.NewReader("not an image"), &output, PNG); err == nil {
		t.Fatal("ConvertStream() error = nil, want decode error")
	}
}

func TestConvertStreamRejectsWebPEncoding(t *testing.T) {
	var input bytes.Buffer
	if err := png.Encode(&input, testImage()); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := ConvertStream(&input, &output, WebP); !errors.Is(err, ErrWebPEncode) {
		t.Fatalf("ConvertStream() error = %v, want ErrWebPEncode", err)
	}
}

func TestConvertStreamRejectsTooManyPixels(t *testing.T) {
	input, err := os.Open("../testdata/oversized.png")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()

	var output bytes.Buffer
	if err := ConvertStream(input, &output, PNG); !errors.Is(err, ErrTooManyPixels) {
		t.Fatalf("ConvertStream() error = %v, want ErrTooManyPixels", err)
	}
	if output.Len() != 0 {
		t.Errorf("wrote %d bytes for an oversized image", output.Len())
	}
}

func TestConvertStreamLimitDisabled(t *testing.T) {
	var input bytes.Buffer
	if err := png.Encode(&input, testImage()); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := ConvertStreamLimit(&input, &output, PNG, 0); err != nil {
		t.Fatalf("ConvertStreamLimit() error: %v", err)
	}
}

func TestToPNG(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.jpg")
	output := filepath.Join(dir, "output.png")

	writeImage(t, input, "jpeg", testImage())

	if err := ToPNG(input, output); err != nil {
		t.Fatalf("ToPNG() error: %v", err)
	}
	assertFormat(t, output, "png", 2, 2)
}

func TestToJPEG(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	output := filepath.Join(dir, "output.jpg")

	writeImage(t, input, "png", testImage())

	if err := ToJPEG(input, output); err != nil {
		t.Fatalf("ToJPEG() error: %v", err)
	}
	assertFormat(t, output, "jpeg", 2, 2)
}

func TestConvertUsesOutputExtension(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "output.png")

	if err := Convert(webpFixture, output); err != nil {
		t.Fatalf("Convert() error: %v", err)
	}
	assertFormat(t, output, "png", 4, 4)
}

func TestConvertRejectsUnknownOutputExtension(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "output.bmp")

	err := Convert(webpFixture, output)
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("Convert() error = %v, want ErrUnsupportedFormat", err)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Errorf("output file was created for unsupported format")
	}
}

func TestToWebPReportsUnsupportedEncoding(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	output := filepath.Join(dir, "output.webp")

	writeImage(t, input, "png", testImage())

	if err := ToWebP(input, output); !errors.Is(err, ErrWebPEncode) {
		t.Fatalf("ToWebP() error = %v, want ErrWebPEncode", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Errorf("output file was created after failed encoding")
	}
}

func TestConvertSkipsOutputOnDecodeFailure(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "missing.png")
	output := filepath.Join(dir, "output.jpg")

	writeFile(t, input, []byte("not an image"))

	if err := ToJPEG(input, output); err == nil {
		t.Fatal("ToJPEG() error = nil, want decode error")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Errorf("output file was created after failed decode")
	}
}

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	return img
}

func writeImage(t *testing.T, path, format string, img image.Image) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	switch format {
	case "jpeg":
		err = jpeg.Encode(f, img, nil)
	case "png":
		err = png.Encode(f, img)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFormat(t *testing.T, path, want string, width, height int) {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if format != want {
		t.Errorf("format = %q, want %q", format, want)
	}
	if got := img.Bounds().Dx(); got != width {
		t.Errorf("width = %d, want %d", got, width)
	}
	if got := img.Bounds().Dy(); got != height {
		t.Errorf("height = %d, want %d", got, height)
	}
}

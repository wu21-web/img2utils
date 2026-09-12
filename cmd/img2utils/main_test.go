package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestRunConvert(t *testing.T) {
	input := encodePNG(t)

	var output bytes.Buffer
	if code := run([]string{"convert", "--to", "jpg"}, input, &output); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
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

func TestRunConvertAcceptsFormatName(t *testing.T) {
	for _, format := range []string{"png", "PNG", ".png"} {
		input := encodePNG(t)

		var output bytes.Buffer
		if code := run([]string{"convert", "--to", format}, input, &output); code != 0 {
			t.Fatalf("run(--to %s) = %d, want 0", format, code)
		}
		if _, got, err := image.Decode(&output); err != nil {
			t.Fatal(err)
		} else if got != "png" {
			t.Errorf("format = %q, want png", got)
		}
	}
}

func TestRunConvertRejectsMissingFormat(t *testing.T) {
	if code := run([]string{"convert"}, bytes.NewReader(nil), &bytes.Buffer{}); code != 2 {
		t.Errorf("run() = %d, want 2", code)
	}
}

func TestRunConvertRejectsUnknownFormat(t *testing.T) {
	input := encodePNG(t)

	if code := run([]string{"convert", "--to", "bmp"}, input, &bytes.Buffer{}); code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
}

func TestRunConvertRejectsExtraArguments(t *testing.T) {
	if code := run([]string{"convert", "--to", "png", "extra"}, bytes.NewReader(nil), &bytes.Buffer{}); code != 2 {
		t.Errorf("run() = %d, want 2", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if code := run([]string{"nope"}, bytes.NewReader(nil), &bytes.Buffer{}); code != 2 {
		t.Errorf("run() = %d, want 2", code)
	}
}

func TestRunHelp(t *testing.T) {
	if code := run([]string{"--help"}, bytes.NewReader(nil), &bytes.Buffer{}); code != 0 {
		t.Errorf("run() = %d, want 0", code)
	}
}

func encodePNG(t *testing.T) *bytes.Reader {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 1, color.RGBA{B: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buf.Bytes())
}

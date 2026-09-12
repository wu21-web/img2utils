package cli

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wu21-web/img2utils/internal/convert"
)

const webpFixture = "../testdata/sample.webp"

func TestRunDerivesOutputName(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "hello.png")
	writePNG(t, input)

	var stderr bytes.Buffer
	if code := Run("png2jpg", convert.JPEG, []string{input}, &stderr); code != 0 {
		t.Fatalf("Run() = %d, stderr: %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.jpg")); err != nil {
		t.Errorf("expected hello.jpg: %v", err)
	}
}

func TestRunAcceptsOutputFlag(t *testing.T) {
	for _, flagName := range []string{"-o", "-O"} {
		t.Run(flagName, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "hello.png")
			output := filepath.Join(dir, "custom.jpg")
			writePNG(t, input)

			var stderr bytes.Buffer
			if code := Run("png2jpg", convert.JPEG, []string{input, flagName, output}, &stderr); code != 0 {
				t.Fatalf("Run() = %d, stderr: %s", code, stderr.String())
			}
			if _, err := os.Stat(output); err != nil {
				t.Errorf("expected %s: %v", output, err)
			}
		})
	}
}

func TestRunConvertsWebP(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "hello.png")

	var stderr bytes.Buffer
	code := Run("webp2png", convert.PNG, []string{webpFixture, "-o", output}, &stderr)
	if code != 0 {
		t.Fatalf("Run() = %d, stderr: %s", code, stderr.String())
	}
	assertPNG(t, output)
}

func TestRunRequiresOneInput(t *testing.T) {
	var stderr bytes.Buffer
	if code := Run("png2jpg", convert.JPEG, nil, &stderr); code != 2 {
		t.Fatalf("Run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: png2jpg") {
		t.Errorf("stderr = %q, want usage", stderr.String())
	}
}

func TestRunReportsConversionFailure(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "missing.png")

	var stderr bytes.Buffer
	if code := Run("png2jpg", convert.JPEG, []string{input}, &stderr); code != 1 {
		t.Fatalf("Run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "png2jpg:") {
		t.Errorf("stderr = %q, want command name", stderr.String())
	}
}

func TestRunRejectsUnknownTarget(t *testing.T) {
	var stderr bytes.Buffer
	if code := Run("bmp2png", convert.Format("bmp"), []string{"input.bmp"}, &stderr); code != 1 {
		t.Fatalf("Run() = %d, want 1", code)
	}
}

func writePNG(t *testing.T, path string) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func assertPNG(t *testing.T, path string) {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, format, err := image.Decode(f); err != nil {
		t.Fatal(err)
	} else if format != "png" {
		t.Errorf("format = %q, want png", format)
	}
}

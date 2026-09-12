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

func TestRunConvertsMultipleInputs(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "a.png")
	second := filepath.Join(dir, "b.png")
	writePNG(t, first)
	writePNG(t, second)

	var stderr bytes.Buffer
	if code := Run("png2jpg", convert.JPEG, []string{first, second}, &stderr); code != 0 {
		t.Fatalf("Run() = %d, stderr: %s", code, stderr.String())
	}
	for _, name := range []string{"a.jpg", "b.jpg"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s: %v", name, err)
		}
	}
}

func TestRunExpandsGlob(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "a.png"))
	writePNG(t, filepath.Join(dir, "b.png"))
	if err := os.Mkdir(filepath.Join(dir, "nested.png"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	var stderr bytes.Buffer
	if code := Run("png2jpg", convert.JPEG, []string{"*.png"}, &stderr); code != 0 {
		t.Fatalf("Run() = %d, stderr: %s", code, stderr.String())
	}
	for _, name := range []string{"a.jpg", "b.jpg"} {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("expected %s: %v", name, err)
		}
	}
}

func TestRunExpandsGlobWithNoMatches(t *testing.T) {
	t.Chdir(t.TempDir())

	var stderr bytes.Buffer
	if code := Run("png2jpg", convert.JPEG, []string{"*.png"}, &stderr); code != 1 {
		t.Fatalf("Run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "*.png") {
		t.Errorf("stderr = %q, want the pattern", stderr.String())
	}
}

func TestRunRejectsOutputFlagWithMultipleInputs(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "a.png")
	second := filepath.Join(dir, "b.png")
	output := filepath.Join(dir, "custom.jpg")
	writePNG(t, first)
	writePNG(t, second)

	var stderr bytes.Buffer
	code := Run("png2jpg", convert.JPEG, []string{first, second, "-o", output}, &stderr)
	if code != 2 {
		t.Fatalf("Run() = %d, want 2", code)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Errorf("output file was created")
	}
}

func TestRunReportsEachFailureAndContinues(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.png")
	missing := filepath.Join(dir, "missing.png")
	writePNG(t, good)

	var stderr bytes.Buffer
	if code := Run("png2jpg", convert.JPEG, []string{good, missing}, &stderr); code != 1 {
		t.Fatalf("Run() = %d, want 1", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "good.jpg")); err != nil {
		t.Errorf("expected good.jpg: %v", err)
	}
	if !strings.Contains(stderr.String(), "missing.png") {
		t.Errorf("stderr = %q, want the failing input", stderr.String())
	}
}

func TestRunRejectsUnknownTarget(t *testing.T) {
	var stderr bytes.Buffer
	if code := Run("avif2png", convert.Format("avif"), []string{"input.avif"}, &stderr); code != 1 {
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

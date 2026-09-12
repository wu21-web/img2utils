package cli

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wu21-web/img2utils/internal/convert"
)

func TestMainConvertsForEveryCommand(t *testing.T) {
	for name, target := range commands {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "photo.webp")
			raw, err := os.ReadFile(webpFixture)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(input, raw, 0o644); err != nil {
				t.Fatal(err)
			}

			var stderr bytes.Buffer
			if code := Main([]string{name, input}, &stderr); code != 0 {
				t.Fatalf("Main() = %d, stderr: %s", code, stderr.String())
			}

			f, err := os.Open(filepath.Join(dir, "photo"+target.Ext()))
			if err != nil {
				t.Fatalf("expected output: %v", err)
			}
			defer f.Close()

			_, format, err := image.Decode(f)
			if err != nil {
				t.Fatal(err)
			}
			got, err := convert.ParseFormat(format)
			if err != nil {
				t.Fatalf("decoded format %q: %v", format, err)
			}
			if got != target {
				t.Errorf("command %s wrote %s, want %s", name, got, target)
			}
		})
	}
}

func TestMainStripsWindowsExtension(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "photo.png")
	writePNG(t, input)

	var stderr bytes.Buffer
	if code := Main([]string{"png2jpg.exe", input}, &stderr); code != 0 {
		t.Fatalf("Main() = %d, stderr: %s", code, stderr.String())
	}
}

func TestMainRejectsUnknownCommand(t *testing.T) {
	var stderr bytes.Buffer
	if code := Main([]string{"nope", "photo.png"}, &stderr); code != 2 {
		t.Fatalf("Main() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Errorf("stderr = %q, want unknown command", stderr.String())
	}
	if !strings.Contains(stderr.String(), "bmp2png") {
		t.Errorf("stderr = %q, want the command list", stderr.String())
	}
}

func TestMainRequiresProgramName(t *testing.T) {
	if code := Main(nil, &bytes.Buffer{}); code != 2 {
		t.Fatalf("Main() = %d, want 2", code)
	}
}

func TestCommandsMatchCmdPackages(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "cmd"))
	if err != nil {
		t.Fatal(err)
	}

	packages := make(map[string]bool, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		packages[name] = true
		if name == "img2utils" {
			continue
		}
		if _, ok := commands[name]; !ok {
			t.Errorf("cmd/%s has no entry in commands", name)
		}
	}

	for name := range commands {
		if !packages[name] {
			t.Errorf("commands has %s but cmd/%s does not exist", name, name)
		}
	}
}

package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wu21-web/img2utils/internal/convert"
)

var commands = map[string]convert.Format{
	"bmp2jpg":  convert.JPEG,
	"bmp2png":  convert.PNG,
	"jpg2bmp":  convert.BMP,
	"jpg2png":  convert.PNG,
	"png2bmp":  convert.BMP,
	"png2jpg":  convert.JPEG,
	"webp2bmp": convert.BMP,
	"webp2jpg": convert.JPEG,
	"webp2png": convert.PNG,
}

func Main(argv []string, stderr io.Writer) int {
	if len(argv) == 0 {
		return 2
	}

	name := strings.TrimSuffix(filepath.Base(argv[0]), ".exe")
	target, ok := commands[name]
	if !ok {
		fmt.Fprintf(stderr, "unknown command %q\n", name)
		fmt.Fprintf(stderr, "commands: %s\n", strings.Join(commandNames(), ", "))
		return 2
	}

	return Run(name, target, argv[1:], stderr)
}

func commandNames() []string {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

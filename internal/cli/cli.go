package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"img2utils/internal/convert"
)

func Run(name string, target convert.Format, args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)

	var output string
	flags.StringVar(&output, "o", "", "output file")
	flags.StringVar(&output, "O", "", "output file")
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s input.%s [-o output.%s]\n", name, sourceExt(target), targetExt(target))
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}

	input := flags.Arg(0)
	if output == "" {
		output = defaultOutput(input, target)
	}

	convertFn, err := converter(target)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}
	if err := convertFn(input, output); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}
	return 0
}

func defaultOutput(input string, target convert.Format) string {
	ext := filepath.Ext(input)
	if ext == "" {
		return input + targetExt(target)
	}
	return strings.TrimSuffix(input, ext) + targetExt(target)
}

func converter(target convert.Format) (func(string, string) error, error) {
	switch target {
	case convert.JPEG:
		return convert.ToJPEG, nil
	case convert.PNG:
		return convert.ToPNG, nil
	case convert.WebP:
		return convert.ToWebP, nil
	default:
		return nil, fmt.Errorf("%w: %q", convert.ErrUnsupportedFormat, target)
	}
}

func targetExt(format convert.Format) string {
	switch format {
	case convert.JPEG:
		return ".jpg"
	case convert.PNG:
		return ".png"
	case convert.WebP:
		return ".webp"
	default:
		return ""
	}
}

func sourceExt(format convert.Format) string {
	switch format {
	case convert.JPEG:
		return "jpg"
	case convert.PNG:
		return "png"
	case convert.WebP:
		return "webp"
	default:
		return "input"
	}
}

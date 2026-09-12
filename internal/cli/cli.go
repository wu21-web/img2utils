package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/wu21-web/img2utils/internal/convert"
)

func Run(name string, target convert.Format, args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)

	var output string
	flags.StringVar(&output, "o", "", "output file")
	flags.StringVar(&output, "O", "", "output file")
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s input.%s [-o output%s]\n", name, target.Name(), target.Ext())
		flags.PrintDefaults()
	}

	if err := flags.Parse(normalizeArgs(args)); err != nil {
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

func normalizeArgs(args []string) []string {
	flags := make([]string, 0, len(args))
	positional := make([]string, 0, len(args))
	rest := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case rest:
			positional = append(positional, arg)
		case arg == "--":
			rest = true
		case arg == "-o" || arg == "-O":
			flags = append(flags, arg)
			if i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		case strings.HasPrefix(arg, "-o=") || strings.HasPrefix(arg, "-O="):
			flags = append(flags, arg)
		default:
			positional = append(positional, arg)
		}
	}

	return append(flags, positional...)
}

func defaultOutput(input string, target convert.Format) string {
	ext := filepath.Ext(input)
	if ext == "" {
		return input + target.Ext()
	}
	return strings.TrimSuffix(input, ext) + target.Ext()
}

func converter(target convert.Format) (func(string, string) error, error) {
	switch target {
	case convert.JPEG:
		return convert.ToJPEG, nil
	case convert.PNG:
		return convert.ToPNG, nil
	case convert.BMP:
		return convert.ToBMP, nil
	case convert.WebP:
		return convert.ToWebP, nil
	default:
		return nil, fmt.Errorf("%w: %q", convert.ErrUnsupportedFormat, target)
	}
}

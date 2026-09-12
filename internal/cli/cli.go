package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
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
		fmt.Fprintf(stderr, "usage: %s image... [-o output%s]\n", name, target.Ext())
		flags.PrintDefaults()
	}

	if err := flags.Parse(normalizeArgs(args)); err != nil {
		return 2
	}
	inputs, err := expand(flags.Args())
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 2
	}
	if len(inputs) == 0 {
		flags.Usage()
		return 2
	}
	if output != "" && len(inputs) > 1 {
		fmt.Fprintf(stderr, "%s: -o takes a single input\n", name)
		return 2
	}

	outputs, err := outputPaths(inputs, output, target)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 2
	}

	convertFn, err := converter(target)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}

	failed := false
	for i, input := range inputs {
		if err := convertFn(input, outputs[i]); err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", name, err)
			failed = true
		}
	}
	if failed {
		return 1
	}
	return 0
}

func outputPaths(inputs []string, output string, target convert.Format) ([]string, error) {
	sources := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		sources[filepath.Clean(input)] = struct{}{}
	}

	outputs := make([]string, len(inputs))
	owners := make(map[string]string, len(inputs))

	for i, input := range inputs {
		path := output
		if path == "" {
			path = defaultOutput(input, target)
		}
		path = filepath.Clean(path)

		if owner, ok := owners[path]; ok {
			return nil, fmt.Errorf("%s and %s both write to %s", owner, input, path)
		}
		if _, ok := sources[path]; ok {
			return nil, fmt.Errorf("refusing to overwrite input %s", path)
		}

		owners[path] = input
		outputs[i] = path
	}

	return outputs, nil
}

func expand(args []string) ([]string, error) {
	expanded := make([]string, 0, len(args))

	for _, arg := range args {
		if _, err := os.Stat(arg); err == nil || !strings.ContainsAny(arg, "*?[") {
			expanded = append(expanded, arg)
			continue
		}

		matches, err := filepath.Glob(arg)
		if err != nil {
			return nil, err
		}

		files := make([]string, 0, len(matches))
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && !info.IsDir() {
				files = append(files, match)
			}
		}
		if len(files) == 0 {
			expanded = append(expanded, arg)
			continue
		}
		expanded = append(expanded, files...)
	}

	return expanded, nil
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

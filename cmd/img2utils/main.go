package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/wu21-web/img2utils/internal/convert"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout))
}

func run(args []string, stdin io.Reader, stdout io.Writer) int {
	if len(args) == 0 {
		usage()
		return 2
	}

	switch args[0] {
	case "convert":
		return runConvert(args[1:], stdin, stdout)
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "img2utils: unknown command %q\n", args[0])
		usage()
		return 2
	}
}

func runConvert(args []string, stdin io.Reader, stdout io.Writer) int {
	flags := flag.NewFlagSet("convert", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	to := flags.String("to", "", "target format: jpg, png, or webp")
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: img2utils convert --to <format> < input > output")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *to == "" || flags.NArg() != 0 {
		flags.Usage()
		return 2
	}

	format, err := convert.ParseFormat(*to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "img2utils: %v\n", err)
		return 1
	}
	if err := convert.ConvertStream(stdin, stdout, format); err != nil {
		fmt.Fprintf(os.Stderr, "img2utils: %v\n", err)
		return 1
	}
	return 0
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: img2utils <command>")
	fmt.Fprintln(os.Stderr, "commands: convert")
}

package main

import (
	"os"

	"github.com/wu21-web/img2utils/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args, os.Stderr))
}

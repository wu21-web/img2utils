package main

import (
	"os"

	"github.com/wu21-web/img2utils/internal/cli"
	"github.com/wu21-web/img2utils/internal/convert"
)

func main() {
	os.Exit(cli.Run("webp2png", convert.PNG, os.Args[1:], os.Stderr))
}

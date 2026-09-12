package main

import (
	"os"

	"img2utils/internal/cli"
	"img2utils/internal/convert"
)

func main() {
	os.Exit(cli.Run("webp2png", convert.PNG, os.Args[1:], os.Stderr))
}

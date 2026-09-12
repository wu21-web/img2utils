package main

import (
	"os"

	"img2utils/internal/cli"
	"img2utils/internal/convert"
)

func main() {
	os.Exit(cli.Run("png2jpg", convert.JPEG, os.Args[1:], os.Stderr))
}

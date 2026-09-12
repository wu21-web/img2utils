package main

import (
	"os"

	"github.com/wu21-web/img2utils/internal/cli"
	"github.com/wu21-web/img2utils/internal/convert"
)

func main() {
	os.Exit(cli.Run("png2jpg", convert.JPEG, os.Args[1:], os.Stderr))
}

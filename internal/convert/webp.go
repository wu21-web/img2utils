package convert

import (
	"errors"

	_ "golang.org/x/image/webp"
)

var ErrWebPEncode = errors.New("webp encoding is not supported")

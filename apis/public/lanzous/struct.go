package lanzous

import (
	"io"
)

type wssOptions struct {
	token string
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

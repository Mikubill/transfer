package tmplink

import (
	"io"
)

type wssOptions struct {
	token     string
	DebugMode bool
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

package whc

import (
	"io"
)

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64

	uploadName string
	password   string
	delete     string
	fileid     int64
}

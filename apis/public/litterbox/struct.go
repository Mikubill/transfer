package litterbox

import (
	"io"
)

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

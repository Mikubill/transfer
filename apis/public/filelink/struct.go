package filelink

import "io"

type cbOptions struct {
	DebugMode bool
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

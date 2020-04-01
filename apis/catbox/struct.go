package catbox

import "github.com/cheggaaa/pb/v3"

type cbOptions struct {
	parallel  int
	prefix    string
	forceMode bool
	debugMode bool
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader *pb.Reader
	fileSize   int64
}

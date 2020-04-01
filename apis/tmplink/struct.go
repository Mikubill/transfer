package tmplink

import "github.com/cheggaaa/pb/v3"

type wssOptions struct {
	token     string
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

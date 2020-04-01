package firefox

import "github.com/cheggaaa/pb/v3"

type ffOptions struct {
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

type keySuite struct {
	nonce      string
	secretKey  []byte
	encryptKey []byte
	encryptIV  []byte
	authKey    []byte
	metaKey    []byte
	metaIV     []byte
}

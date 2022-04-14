package methods

import (
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

type MultiPartUploadConfig struct {
	Debug      bool
	Endpoint   string
	FileName   string
	FileReader io.Reader
	FileSize   int64
}

type TransferConfig struct {
	Parallel   int
	DebugMode  *bool
	NoBarMode  bool
	CryptoMode bool
	CryptoKey  string
}

type DownloaderConfig struct {
	TransferConfig
	Link        string
	Prefix      string
	ForceMode   bool
	Modifier    func(r *http.Request)
	RespHandler func(r *http.Response) bool
}

type parallelConfig struct {
	parallel int
	modifier func(r *http.Request)
	counter  *writeCounter
	wg       *sync.WaitGroup
}

type writeCounter struct {
	bar    *pb.ProgressBar
	offset int64
	writer *os.File
}

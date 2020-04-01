package utils

import (
	"github.com/cheggaaa/pb/v3"
	"net/http"
	"os"
	"sync"
)

type Backend interface {
	Upload([]string)
	Download([]string)
	GetArgs() [][]string
	SetArgs()
}

type MainConfig struct {
	Commands [][]string
	Backend  Backend
	KeepMode bool
	Version  bool
}

type DownloadConfig struct {
	Force    bool
	Debug    bool
	Parallel int
	Modifier func(r *http.Request)
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

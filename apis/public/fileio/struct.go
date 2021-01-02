package fileio

import (
	"io"
)

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

type uploadResp struct {
	Success bool   `json:"success"`
	Key     string `json:"key"`
	Link    string `json:"link"`
}

package gofile

import "github.com/cheggaaa/pb/v3"

type goFileOptions struct {
	token     string
	parallel  int
	prefix    string
	forceMode bool
	debugMode bool
}

type respBody struct {
	Status string    `json:"status"`
	Data   respBlock `json:"data"`
}

type respBlock struct {
	Server      string `json:"server"`
	Code        string `json:"code"`
	RemovalCode string `json:"removalCode"`
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader *pb.Reader
	fileSize   int64
}

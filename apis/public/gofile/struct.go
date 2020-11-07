package gofile

import "github.com/cheggaaa/pb/v3"

type goFileOptions struct {
	token      string
	singleMode bool
}

type respBody struct {
	Status string    `json:"status"`
	Data   respBlock `json:"data"`
}

type respBlock struct {
	Server      string     `json:"server"`
	Code        string     `json:"code"`
	RemovalCode string     `json:"adminCode"`
	Items       map[string]fileItem `json:"files"`
}

type fileItem struct {
	Name string `json:"name"`
	Size int64 `json:"size"`
	Link string `json:"link"`
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader *pb.Reader
	fileSize   int64
}

package bitsend

import (
	"io"
)

type wssOptions struct {
	interval int
	passCode string
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

type uploadResp struct {
	Files []uploadRespBlock `json:"files"`
}

type uploadRespBlock struct {
	Name       string `json:"name"`
	Size       int    `json:"size"`
	Type       string `json:"type"`
	NiceSize   string `json:"niceSize"`
	RealName   string `json:"realName"`
	FileKey    string `json:"fileKey"`
	DelFileKey string `json:"delFileKey"`
	DeleteUrl  string `json:"delete_url"`
	DeleteType string `json:"delete_type"`
}

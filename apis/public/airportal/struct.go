package airportal

import (
	"io"
)

type arpOptions struct {
	token     string
	downloads int
	hours     int
	username  string
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
	ticket     uploadTicket
}

type uploadTicket struct {
	Code      int    `json:"code"`
	Accessid  string `json:"accessid"`
	Host      string `json:"host"`
	Key       string `json:"key"`
	Policy    string `json:"policy"`
	Signature string `json:"signature"`
	Alert     string `json:"alert"`
}

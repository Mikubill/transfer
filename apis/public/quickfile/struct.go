package quickfile

import (
	"io"
	"time"
)

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

type uploadResp struct {
	ID          string `json:"Id"`
	UID         string `json:"Uid"`
	FileName    string `json:"FileName"`
	StorageName string `json:"StorageName"`
	Provider    struct {
		Name string `json:"Name"`
		URL  string `json:"Url"`
		Free bool   `json:"Free"`
	} `json:"Provider"`
	Md5         string    `json:"Md5"`
	Path        string    `json:"Path"`
	Size        int       `json:"Size"`
	ContentType string    `json:"ContentType"`
	URL         string    `json:"Url"`
	ShortURL    string    `json:"ShortUrl"`
	QrCode      string    `json:"QrCode"`
	LinkID      string    `json:"LinkId"`
	AutoDelete  bool      `json:"AutoDelete"`
	Expiration  int       `json:"Expiration"`
	ExpiresAt   time.Time `json:"ExpiresAt"`
	Time        time.Time `json:"Time"`
}

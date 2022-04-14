package tmplink

import (
	"io"
)

type tmpOptions struct {
	token   string
	dest    string
	anom    bool
	upToken string
}

type tmpEpResp struct {
	Data struct {
		Utoken   string `json:"utoken"`
		Uploader string `json:"uploader"`
		Src      string `json:"src"`
	} `json:"data"`
	Status int   `json:"status"`
	Debug  []any `json:"debug"`
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

type tmpSignupResp struct {
	Data   string `json:"data"`
	Status int    `json:"status"`
	Debug  []any  `json:"debug"`
}

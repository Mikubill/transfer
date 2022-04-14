package anonfiles

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
	Status bool `json:"status"`
	Data   struct {
		File struct {
			URL struct {
				Full  string `json:"full"`
				Short string `json:"short"`
			} `json:"url"`
			Metadata struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Size struct {
					Bytes    int    `json:"bytes"`
					Readable string `json:"readable"`
				} `json:"size"`
			} `json:"metadata"`
		} `json:"file"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

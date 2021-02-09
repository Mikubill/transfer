package wenshushu

import (
	"net/http"
	"sync"
	"time"
)

type wssOptions struct {
	token      string
	interval   int
	singleMode bool
	blockSize  int
	passCode   string
	Parallel   int
}

type requestConfig struct {
	debug    bool
	retry    int
	timeout  time.Duration
	modifier func(r *http.Request)
}

type uploadPart struct {
	content []byte
	count   int64
	name    string
	wg      *sync.WaitGroup
}

type sendConfigResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    sendConfigBlock `json:"data"`
}

type timeConfigResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Exp  int    `json:"exp"`
		Time string `json:"time"`
	} `json:"data"`
}

type sendConfigBlock struct {
	Bid         string     `json:"bid"`
	SocialToken string     `json:"social_token"`
	Tid         string     `json:"tid"`
	UFileID     string     `json:"ufileid"`
	UploadID    string     `json:"upId"`
	URL         string     `json:"url"`
	P           int        `json:"pro"`
	R           string     `json:"rst"`
	ManageURL   string     `json:"mgr_url"`
	PublicURL   string     `json:"public_url"`
	SocialURL   string     `json:"social_url"`
	Token       string     `json:"token"`
	BoxID       string     `json:"boxid"`
	Createdat   string     `json:"createdat"`
	Expire      string     `json:"expire"`
	FileCount   string     `json:"file_count"`
	FileSize    string     `json:"file_size"`
	TaskID      string     `json:"taskid"`
	FileList    []fileItem `json:"fileList"`
}

type fileItem struct {
	Bid      string `json:"bid"`
	Fid      string `json:"fid"`
	FileName string `json:"fname"`
}

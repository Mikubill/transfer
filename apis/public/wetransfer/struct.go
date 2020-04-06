package wetransfer

import (
	"net/http"
	"sync"
	"time"
)

type wetOptions struct {
	interval   int
	singleMode bool
	Parallel   int
}

type requestConfig struct {
	action   string
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
	fileID  string
}

type requestTicket struct {
	token   string
	cookies string
}

type fileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"item_type"`
}

type configBlock struct {
	ID       string `json:"id"`
	State    string `json:"state"`
	Hash     string `json:"security_hash"`
	ticket   requestTicket
	URL      string     `json:"url"`
	Public   string     `json:"shortened_url"`
	Item     []fileInfo `json:"items"`
	Download string     `json:"direct_link"`
}

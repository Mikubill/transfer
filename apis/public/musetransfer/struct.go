package musetransfer

import (
	"net/http"
	"sync"
	"time"
)

type finResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  any    `json:"result"`
	TraceID any    `json:"traceId"`
}

type addEntryResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  struct {
		ID int `json:"id"`
	} `json:"result"`
	TraceID any `json:"traceId"`
}

type createResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  struct {
		Code       string `json:"code"`
		UploadPath string `json:"uploadPath"`
	} `json:"result"`
	TraceID any `json:"traceId"`
}

type museOptions struct {
	interval   int
	singleMode bool
	Parallel   int

	dest        string
	sToken      string
	devicetoken string
}

type fsListResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  []struct {
		Code          string `json:"code"`
		FileID        string `json:"fileId"`
		FileType      int    `json:"fileType"`
		Name          string `json:"name"`
		Folder        any    `json:"folder"`
		Size          int    `json:"size"`
		Type          string `json:"type"`
		PreviewStatus int    `json:"previewStatus"`
		ParentFileID  string `json:"parentFileId"`
		Count         any    `json:"count"`
		Thumbnail     any    `json:"thumbnail"`
		Preview       any    `json:"preview"`
	} `json:"result"`
	TraceID any `json:"traceId"`
}

type dlResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  struct {
		URL string `json:"url"`
	} `json:"result"`
	TraceID any `json:"traceId"`
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
	dest    string
	auth    *s3Token
}

type s3Token struct {
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	SecurityToken   string `json:"securityToken"`
}

type getTokenResp struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Result  s3Token `json:"result"`
	TraceID any     `json:"traceId"`
}

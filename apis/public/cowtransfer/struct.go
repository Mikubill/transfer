package cowtransfer

import (
	"net/http"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
)

type requestConfig struct {
	debug  bool
	action string
	//retry    int
	timeout  time.Duration
	modifier func(r *http.Request)
}

type uploadPart struct {
	content []byte
	count   int64
	bar     *pb.ProgressBar
}

type uploadConfig struct {
	wg      *sync.WaitGroup
	config  *initResp
	hashMap *cmap.ConcurrentMap
}

type cowOptions struct {
	Parallel   int
	token      string
	authCode   string
	interval   int
	singleMode bool
	blockSize  int64
	hashCheck  bool
	passCode   string
	validDays  int
}

type initResp struct {
	Token        string
	TransferGUID string
	FileGUID     string
	EncodeID     string
	Exp          int64  `json:"expireAt"`
	ID           string `json:"uploadId"`
}

type prepareSendResp struct {
	UploadToken  string `json:"uptoken"`
	TransferGUID string `json:"transferguid"`
	FileGUID     string `json:"fileguid"`
	UniqueURL    string `json:"uniqueurl"`
	Prefix       string `json:"prefix"`
	QRCode       string `json:"qrcode"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
}

type beforeSendResp struct {
	FileGuid string `json:"fileGuid"`
}

type uploadResponse struct {
	Error string `json:"error"`
	Etag  string `json:"etag"`
	MD5   string `json:"md5"`
}

type slek struct {
	ETag string `json:"etag"`
	Part int64  `json:"partNumber"`
}

type clds struct {
	Parts    []slek `json:"parts"`
	FName    string `json:"fname"`
	Mimetype string `json:"mimeType"`
	Metadata map[string]string
	Vars     map[string]string
}

type finishResponse struct {
	TempDownloadCode string `json:"tempDownloadCode"`
	Status           bool   `json:"complete"`
}

type downloadDetailsResponse struct {
	GUID         string `json:"guid"`
	DownloadName string `json:"downloadName"`
	Deleted      bool   `json:"deleted"`
	Uploaded     bool   `json:"uploaded"`
	// Details      []downloadDetailsBlock `json:"transferFileDtos"`
}

type downloadFilesResponse struct {
	Details []downloadDetailsBlock `json:"files"`
}

type downloadDetailsBlock struct {
	GUID     string  `json:"guid"`
	FileName string  `json:"fileName"`
	Size     float64 `json:"sizeInByte"`
}

type uploadResult struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

type downloadConfigResponse struct {
	Link string `json:"link"`
}

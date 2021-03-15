package notion

import (
	"io"

	"github.com/kjk/notionapi"
)

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

// POST /api/v3/getUploadFileUrl request
type getUploadFileUrlRequest struct {
	Bucket      string `json:"bucket"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

// GetUploadFileUrlResponse is a response to POST /api/v3/getUploadFileUrl
type GetUploadFileUrlResponse struct {
	URL          string `json:"url"`
	SignedGetURL string `json:"signedGetUrl"`
	SignedPutURL string `json:"signedPutUrl"`

	FileID string `json:"-"`

	RawJSON map[string]interface{} `json:"-"`
}

type Client struct {
	notionapi.Client
}

type submitTransactionRequest struct {
	RequestID   string        `json:"requestId"`
	Transaction []Transaction `json:"transactions"`
}

type Transaction struct {
	ID         string       `json:"id"`
	SpaceID    string       `json:"spaceId"`
	Operations []*Operation `json:"operations"`
}

type Operation struct {
	Point   Pointer     `json:"pointer"`
	Path    []string    `json:"path"`
	Command string      `json:"command"`
	Args    interface{} `json:"args"`
}

type Pointer struct {
	ID    string `json:"id"`
	Table string `json:"table"`
}

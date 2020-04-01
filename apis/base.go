package apis

import (
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

type BaseBackend interface {
	Uploader
	Downloader
	SetArgs(*cobra.Command)
	LinkMatcher(string) bool
}

type DownloaderConfig struct {
	Link     string
	Config   DownConfig
	Modifier func(r *http.Request)
}

type DownConfig struct {
	Prefix    string
	DebugMode bool
	ForceMode bool
	Ticket    string
	Parallel  int
}

type Uploader interface {
	InitUpload([]string, []int64) error
	PreUpload(string, int64) error
	DoUpload(string, int64, io.Reader) error
	PostUpload(string, int64) error
	FinishUpload([]string) error
}

type Downloader interface {
	DoDownload(string, DownConfig) error
}

type Backend struct {
	BaseBackend
}

func (b Backend) InitUpload([]string, []int64) error {
	return nil
}

func (b Backend) FinishUpload([]string) error {
	return nil
}

func (b Backend) PreUpload(string, int64) error {
	return nil
}

func (b Backend) DoUpload(string, int64, io.Reader) error {
	panic("method DoUpload is not implemented")
}

func (b Backend) PostUpload(string, int64) error {
	return nil
}

func (b Backend) DoDownload(link string, config DownConfig) error {
	return DownloadFile(&DownloaderConfig{
		Link:     link,
		Config:   config,
		Modifier: AddHeaders,
	})
}

func AddHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 Transfer/0.1.36")
	req.Header.Set("Origin", req.Host)
	req.Header.Set("Referer", req.Host)
}

package apis

import (
	"io"
	"net/http"

	"github.com/Mikubill/transfer/apis/methods"
	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

type BaseBackend interface {
	Uploader
	Downloader
	SetArgs(*cobra.Command)
	LinkMatcher(string) bool
}

type DownConfig struct {
	methods.DownloaderConfig
	Ticket string
}

type Uploader interface {
	InitUpload([]string, []int64) error
	PreUpload(string, int64) error
	DoUpload(string, int64, io.Reader) error
	PostUpload(string, int64) (string, error)
	FinishUpload([]string) (string, error)

	StartProgress(io.Reader, int64) io.Reader
	EndProgress()
}

type Downloader interface {
	DoDownload(string, DownConfig) error
}

type Backend struct {
	BaseBackend
	Bar *pb.ProgressBar
}

func (b *Backend) StartProgress(stream io.Reader, size int64) io.Reader {
	bar := pb.Full.Start64(size)
	reader := bar.NewProxyReader(stream)
	b.Bar = bar
	return reader
}

func (b Backend) EndProgress() {
	b.Bar.Finish()
}

func (b Backend) InitUpload([]string, []int64) error {
	return nil
}

func (b Backend) FinishUpload([]string) (string, error) {
	return "", nil
}

func (b Backend) PreUpload(string, int64) error {
	return nil
}

func (b Backend) DoUpload(string, int64, io.Reader) error {
	panic("method DoUpload is not implemented")
}

func (b Backend) PostUpload(string, int64) (string, error) {
	return "", nil
}

func (b Backend) DoDownload(link string, config DownConfig) error {
	config.Link = link
	config.Modifier = AddHeaders
	return DownloadFile(config)
}

func AddHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; U; Linux x86_64; zh-CN; rv:1.9.2.10) "+
		"Gecko/20100922 Ubuntu/10.10 (maverick) Firefox/3.6.10")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;")
	req.Header.Set("Origin", req.Host)
	req.Header.Set("Referer", req.Host)
}

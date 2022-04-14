package image

import (
	"bytes"
	"fmt"
	"github.com/Mikubill/transfer/utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
)

type PicBed interface {
	Upload([]byte) (string, error)
	UploadStream(chan UploadDataFlow)     // optional
	DownloadStream(chan DownloadDataFlow) // optional
}

type picBed struct {
	PicBed
}

type mWriter struct {
	*multipart.Writer
}

type UploadDataFlow struct {
	Wg      *sync.WaitGroup
	Data    []byte
	Offset  int64
	HashMap *cmap.ConcurrentMap
}

type DownloadDataFlow struct {
	Wg     *sync.WaitGroup
	File   *os.File
	Bar    *pb.ProgressBar
	Hash   string
	Offset string
}

// func (s *picBed) linkBuilder(string) string {
// 	panic("linkBuilder method not implemented")
// }

func (s picBed) Upload([]byte) (string, error) {
	panic("Upload method not implemented")
}

func (s picBed) UploadStream(chan UploadDataFlow) {
	panic("UploadStream method not implemented")
}

func (s picBed) DownloadStream(chan DownloadDataFlow) {
	panic("DownloadStream method not implemented")
}

func defaultReqMod(req *http.Request) {
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:65.0) Gecko/20100101 Firefox/65.0")
	req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("accept-language", "en-US,en;q=0.5")
	req.Header.Set("origin", req.URL.String())
	req.Header.Set("referer", req.URL.String())
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("x-csrf-token", "")
}

func (s picBed) upload(data []byte, postURL string, fieldName string,
	reqModifier func(*http.Request)) ([]byte, error) {

	if Verbose {
		fmt.Println("requesting: " + postURL)
		fmt.Printf("body: byte(%d)\n", len(data))
	}

	client := http.Client{Timeout: 30 * time.Second}
	byteBuf := &bytes.Buffer{}
	writer := NewWriter(byteBuf)

	filename := utils.GenRandString(14) + ".png"
	w, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, err
	}
	_, _ = w.Write(data)
	_ = writer.Close()
	req, err := http.NewRequest("POST", postURL, byteBuf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	reqModifier(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	if Verbose {
		fmt.Println("returns: " + string(body))
	}

	return body, nil
}

func NewWriter(w io.Writer) *mWriter {
	m := &mWriter{multipart.NewWriter(w)}
	m.SetBoundary("------WebKitFormBoundary" + utils.GenRandString(10))
	return m
}

func (w *mWriter) CreateFormFile(field, file string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(field), escapeQuotes(file)))
	h.Set("Content-Type", "image/png")
	return w.CreatePart(h)
}

func escapeQuotes(i string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(i)
}

package image

import (
	"bytes"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"transfer/utils"
)

type PicBed interface {
	Upload([]byte) (string, error)
	UploadStream(chan UploadDataFlow)
	DownloadStream(chan DownloadDataFlow)
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

func (s *picBed) linkExtractor(string) string {
	panic("linkExtractor method not implemented")
}

func (s *picBed) linkBuilder(string) string {
	panic("linkBuilder method not implemented")
}

func (s picBed) UploadStream(dataChan chan UploadDataFlow) {
	for {
		data, ok := <-dataChan
		if !ok {
			break
		}
		url, err := s.Upload(data.Data)
		if err != nil {
			dataChan <- data
			continue
		}
		data.HashMap.Set(strconv.FormatInt(data.Offset, 10), s.linkExtractor(url))
		data.Wg.Done()
	}
}

func (s picBed) DownloadStream(dataChan chan DownloadDataFlow) {
	for {
		data, ok := <-dataChan
		if !ok {
			break
		}
		link := s.linkBuilder(data.Hash)
		resp, err := http.Get(link)
		if err != nil {
			dataChan <- data
			continue
		}
		bd, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			dataChan <- data
			continue
		}
		_ = resp.Body.Close()
		offset, _ := strconv.ParseInt(data.Offset, 10, 64)
		n, _ := data.File.WriteAt(bd, offset)
		data.Bar.Add(n)
		data.Wg.Done()
	}
}

func (s picBed) upload(data []byte, postURL string, fieldName string) ([]byte, error) {

	if Verbose {
		fmt.Println("requesting: " + postURL)
		fmt.Printf("body: byte(%d)\n", len(data))
	}

	client := http.Client{Timeout: 15 * time.Second}
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
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:65.0) Gecko/20100101 Firefox/65.0")
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
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
	return &mWriter{multipart.NewWriter(w)}
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

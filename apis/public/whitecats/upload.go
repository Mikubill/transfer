package whc

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const upload = "http://whitecats.dip.jp/up/upload/%d"
const download = "http://whitecats.dip.jp/up/download/%s"

var (
	ext        = regexp.MustCompile(`[0-9]{10}`)
	errExtract = regexp.MustCompile(`error_message">(.*)</`)
)

func (b *whiteCats) DoUpload(name string, size int64, file io.Reader) error {

	if b.pwd == "" {
		b.pwd = utils.GenRandString(8)
	}
	if b.del == "" {
		b.del = utils.GenRandString(8)
	}

	rand.Seed(time.Now().UnixNano())
	body, err := b.newMultipartUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      apis.DebugMode,
		uploadName: utils.GenRandString(4) + ".bin",
		password:   b.pwd,
		delete:     b.del,
		fileid:     int64(rand.Int31()),
	})

	if err != nil {
		return fmt.Errorf("upload returns error: %s", err)
	}

	b.resp = ext.FindString(string(body))
	return nil
}

func (b *whiteCats) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: %s\n", fmt.Sprintf(download, b.resp))
	fmt.Printf("Download Password: %s\n", b.pwd)
	fmt.Printf("Remove Code: %s\n", b.del)
	return b.resp, nil
}

func (b whiteCats) newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("comment", config.fileName)
	_ = writer.WriteField("download_pass", config.password)
	_ = writer.WriteField("remove_pass", config.delete)
	_ = writer.WriteField("code_pat", "京")
	_ = writer.WriteField("submit", "ファイルを送信する")

	_, err := writer.CreateFormFile("file", config.uploadName)
	if err != nil {
		return nil, err
	}

	writerLength := byteBuf.Len()
	writerBody := make([]byte, writerLength)
	_, _ = byteBuf.Read(writerBody)
	_ = writer.Close()

	lastBoundary := fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary())
	totalSize := int64(writerLength) + config.fileSize + int64(len(lastBoundary))
	partR, partW := io.Pipe()

	go func() {
		_, _ = partW.Write(writerBody)
		for {
			buf := make([]byte, 256)
			nr, err := io.ReadFull(config.fileReader, buf)
			if nr <= 0 {
				break
			}
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				fmt.Println(err)
				break
			}
			if nr > 0 {
				_, _ = partW.Write(buf[:nr])
			}
		}
		_, _ = fmt.Fprintf(partW, lastBoundary)
		_ = partW.Close()
	}()

	req, err := http.NewRequest("POST", fmt.Sprintf(upload, config.fileid), partR)
	if err != nil {
		return nil, err
	}
	req.ContentLength = totalSize
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Referer", "http://whitecats.dip.jp/up/")

	req.Header.Set("content-length", strconv.FormatInt(totalSize, 10))
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	if config.debug {
		log.Printf("header: %v", req.Header)
	}
	resp, err := client.Do(req)
	if err != nil {
		if config.debug {
			log.Printf("do requests returns error: %v", err)
		}
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", body)
		}
		return nil, err
	}
	_ = resp.Body.Close()
	str, _, _ := transform.String(japanese.EUCJP.NewDecoder(), string(body))
	if config.debug {
		log.Printf("returns: %v", str)
	}
	if p := errExtract.FindStringSubmatch(str); len(p) > 0 {
		return nil, fmt.Errorf(p[1])
	}

	return body, nil
}

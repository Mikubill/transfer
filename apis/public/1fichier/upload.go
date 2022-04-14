package fichier

import (
	"bytes"
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	extractDownload = regexp.MustCompile(`https://1fichier.com/\?\w+`)
	extractRemove   = regexp.MustCompile(`https://1fichier.com/remove/[\w/]+`)
	extractUpload   = regexp.MustCompile(`https://[\w./-]+.1fichier.com/upload.cgi[\w.?/=]+`)
	// extact          = regexp.MustCompile(`[0-9]{10}`)
)

func (b *fichier) DoUpload(name string, size int64, file io.Reader) error {

	req, err := http.NewRequest("GET", "https://1fichier.com/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", fmt.Sprintf("SID=%s;show_cm=no;", b.cookie))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	bs, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	uploadURL := extractUpload.Find(bs)
	// log.Println(string(uploadURL))
	if uploadURL == nil {
		// log.Println(string(bs))
		uploadURL = []byte(fmt.Sprintf("https://up2.1fichier.com/upload.cgi?id=%s", utils.GenRandString(5)))
	}

	rand.Seed(time.Now().UnixNano())
	body, err := b.newMultipartUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      apis.DebugMode,
		password:   b.pwd,
		uploadURL:  strings.TrimSpace(string(uploadURL)),
	})

	if err != nil {
		return fmt.Errorf("upload returns error: %s", err)
	}

	b.resp = extractDownload.FindString(string(body))
	b.remove = extractRemove.FindString(string(body))
	return nil
}

func (b *fichier) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: %s\n", b.resp)
	if b.pwd != "" {
		fmt.Printf("Download Password: %s\n", b.pwd)
	}
	fmt.Printf("Remove Code: %s\n", b.remove)
	return b.resp, nil
}

func (b fichier) newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("send_ssl", "on")
	_ = writer.WriteField("domain", "0")
	_ = writer.WriteField("cdpass", config.password)
	_ = writer.WriteField("submit", "Send")

	_, err := writer.CreateFormFile("file[]", config.fileName)
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

	req, err := http.NewRequest("POST", config.uploadURL, partR)
	if err != nil {
		return nil, err
	}
	req.ContentLength = totalSize
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Referer", "https://1fichier.com/")

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
	// str, _, _ := transform.String(japanese.EUCJP.NewDecoder(), string(body))
	if config.debug {
		log.Printf("returns: %v", string(body))
	}
	return body, nil
}

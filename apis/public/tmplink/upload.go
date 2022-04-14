package tmplink

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

const (
	loginUploadEp = "https://connect.tmp.link/api_v2/cli_uploader"
	anomFileEp    = "https://tun.tmp.link/api_v2/file"
	anomUserInit  = "https://tun.tmp.link/api_v2/token"
)

var linkMatcher = regexp.MustCompile("[a-f0-9]{13}")

func (b *tmpLink) DoUpload(name string, size int64, file io.Reader) error {

	body, err := b.newMultipartUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      apis.DebugMode,
	})
	if err != nil {
		return fmt.Errorf("upload returns error: %s", err)
	}

	b.resp = fmt.Sprintf("https://tmp.link/f/%s", linkMatcher.Find(body))

	return nil
}

func (b *tmpLink) InitUpload([]string, []int64) error {
	if b.Config.token != "" {
		b.Config.dest = loginUploadEp
		b.Config.anom = false
		b.Config.upToken = b.Config.token
		return nil
	}
	b.Config.anom = true
	return b.anomSignup()
}

func (b *tmpLink) PreUpload(_ string, f int64) error {
	if b.Config.anom {
		return b.getUploadEp(f)
	}
	return nil
}

func (b *tmpLink) anomSignup() error {
	qs := map[string]string{
		"action":  "token",
		"captcha": utils.GenRandString(64),
		"token":   "",
	}
	var payload string
	for k, v := range qs {
		payload += fmt.Sprintf("%s=%s&", k, v)
	}
	resp, err := http.Post(anomUserInit, "application/x-www-form-urlencoded", strings.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r tmpSignupResp
	if err = json.Unmarshal(body, &r); err != nil {
		return err
	}
	if r.Data == "" || r.Status != 1 {
		return fmt.Errorf("anomymous signup currently not available")
	}
	b.Config.token = r.Data
	return nil
}

func (b *tmpLink) getUploadEp(f int64) error {
	qs := map[string]string{
		"action":   "upload_request_select",
		"token":    b.Config.token,
		"sync":     "1",
		"captcha":  utils.GenRandString(64),
		"filesize": strconv.FormatInt(f, 10),
	}
	var payload string
	for k, v := range qs {
		payload += fmt.Sprintf("%s=%s&", k, v)
	}
	resp, err := http.Post(anomFileEp, "application/x-www-form-urlencoded", strings.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r tmpEpResp
	if err = json.Unmarshal(body, &r); err != nil {
		return err
	}
	if r.Status != 1 {
		return fmt.Errorf("anomymous signup currently not available")
	}
	b.Config.dest = r.Data.Uploader
	b.Config.upToken = r.Data.Utoken
	return nil
}

func (b *tmpLink) PostUpload(string, int64) (string, error) {
	link := b.resp
	fmt.Printf("Download Link: %s\n", link)
	return link, nil
}

func (b tmpLink) newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("model", "0")
	_ = writer.WriteField("token", b.Config.token)
	_ = writer.WriteField("utoken", b.Config.upToken)
	_ = writer.WriteField("filename", config.fileName)
	_ = writer.WriteField("mr_id", "0")
	// if b.Config.token != "" {
	// 	_ = writer.WriteField("action", "upload")
	// 	_ = writer.WriteField("token", utils.GenRandString(16))
	// }
	// _ = writer.WriteField("u_key", utils.GenRandString(16))
	_, err := writer.CreateFormFile("file", config.fileName)
	if err != nil {
		return nil, err
	}

	writerLength := byteBuf.Len()
	writerBody := make([]byte, writerLength)
	_, _ = byteBuf.Read(writerBody)
	_ = writer.Close()

	boundary := byteBuf.Len()
	lastBoundary := make([]byte, boundary)
	_, _ = byteBuf.Read(lastBoundary)

	totalSize := int64(writerLength) + config.fileSize + int64(boundary)
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
		_, _ = partW.Write(lastBoundary)
		_ = partW.Close()
	}()
	var req *http.Request
	req, err = http.NewRequest("POST", b.Config.dest, partR)

	if err != nil {
		return nil, err
	}
	req.ContentLength = totalSize
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
			log.Printf("read response returns: %v", err)
		}
		return nil, err
	}
	_ = resp.Body.Close()
	if config.debug {
		log.Printf("returns: %v", string(body))
	}

	return body, nil
}

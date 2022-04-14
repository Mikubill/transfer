package airportal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
)

const (
	index  = "https://airportal.cn"
	upload = "https://airportal.cn/backend/airportal/getcode"
)

func (b *airPortal) PreUpload(name string, size int64) error {

	fp := name
	if filepath.Ext(fp) == "" {
		fp = fp + ".bin"
	}

	fmt.Printf("fetching upload tickets..")
	end := utils.DotTicker()
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("chunksize", "104857600")
	_ = writer.WriteField("downloads", strconv.Itoa(b.Config.downloads))
	_ = writer.WriteField("host", "airportal-cn-north.oss-cn-beijing.aliyuncs.com")
	_ = writer.WriteField("hours", strconv.Itoa(b.Config.hours))
	_ = writer.WriteField("info", fmt.Sprintf(`[{"id":"","name":"%s","type":"application/octet-stream","relativePath":"",
"size":%d,"origSize":%d,"loaded":0,"percent":0,"status":1,"lastModifiedDate":"","completeTimestamp":0}]`, fp,
		size, size))
	_ = writer.WriteField("token", b.Config.token)
	_ = writer.WriteField("username", b.Config.username)
	_ = writer.Close()

	if apis.DebugMode {
		log.Printf("postbody: %s", string(byteBuf.Bytes()))
	}
	req, err := http.NewRequest("POST", upload, byteBuf)
	if err != nil {
		return fmt.Errorf("build request error: %v", err)
	}

	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	req.Header.Set("user-agent", "Mozilla/5.0 transfer")
	req.Header.Set("referer", " https://airportal.cn/")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("get ticket error: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if apis.DebugMode {
		log.Printf("ticket: %s", string(body))
	}
	if err != nil {
		return fmt.Errorf("read ticket error: %v", err)
	}
	_ = resp.Body.Close()
	var tk uploadTicket
	if err := json.Unmarshal(body, &tk); err != nil {
		return err
	}
	fmt.Printf("ok\n")
	*end <- struct{}{}

	if tk.Alert != "" {
		return fmt.Errorf(tk.Alert)
	}
	b.token = tk
	return nil
}

func (b airPortal) DoUpload(name string, size int64, stream io.Reader) error {

	fp := name
	if filepath.Ext(name) == "" {
		fp = fp + ".bin"
	}

	_, err := b.newMultipartUpload(uploadConfig{
		fileSize:   size,
		fileName:   fp,
		fileReader: stream,
		debug:      apis.DebugMode,
		ticket:     b.token,
	})
	if err != nil {
		return fmt.Errorf("upload error: %s", err)
	}

	return nil
}

func (b airPortal) PostUpload(string, int64) (string, error) {
	link := fmt.Sprintf("%s/%d", index, b.token.Code)
	fmt.Printf("Download Link: %s\n", link)
	return link, nil
}

func (b airPortal) newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("name", config.fileName)
	_ = writer.WriteField("chunk", "0")
	_ = writer.WriteField("chunks", "1")
	_ = writer.WriteField("policy", config.ticket.Policy)
	_ = writer.WriteField("OSSAccessKeyId", config.ticket.Accessid)
	_ = writer.WriteField("success_action_status", "200")
	_ = writer.WriteField("signature", config.ticket.Signature)
	_ = writer.WriteField("key", fmt.Sprintf("%d/%s/%d/%s", config.ticket.Code, config.ticket.Key, 1, config.fileName))

	_, err := writer.CreateFormFile("file", config.fileName)
	if err != nil {
		return nil, err
	}
	if apis.DebugMode {
		log.Printf("postbody: %s", string(byteBuf.Bytes()))
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

	req, err := http.NewRequest("POST", "https://"+config.ticket.Host, partR)
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
		log.Printf("returns: %v, %s", string(body), resp.Status)
	}

	return body, nil
}

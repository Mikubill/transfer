package gofile

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"transfer/apis"
	"transfer/utils"
)

const (
	getServer = "https://apiv2.gofile.io/getServer"
)

func (b *goFile) InitUpload(_ []string, sizes []int64) error {
	err := b.selectServer()
	if err != nil {
		return err
	}
	if b.Config.singleMode {
		b.initUpload(sizes)
		b.initPipe()
		go b.initMultipartUpload()
		_, _ = b.streamWriter.Write(b.baseBody)
	}
	return nil
}

func (b *goFile) initUpload(sizes []int64) {
	var totalSize int64
	for _, v := range sizes {
		totalSize += v
	}
	b.totalSize = totalSize
	b.dataCh = make(chan []byte)
}

func (b *goFile) selectServer() error {

	fmt.Printf("selecting server..")
	end := utils.DotTicker()
	body, err := http.Get(getServer)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}

	data, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return fmt.Errorf("read body returns error: %v", err)
	}
	_ = body.Body.Close()

	var sevData respBody
	if err := json.Unmarshal(data, &sevData); err != nil {
		return fmt.Errorf("parse body returns error: %v", err)
	}
	*end <- struct{}{}
	fmt.Printf("%s\n", strings.TrimSpace(sevData.Data.Server))
	// srv-store0 has dns problem
	if sevData.Data.Server == "srv-store0" {
		sevData.Data.Server = "srv-store1"
	}
	b.serverLink = fmt.Sprintf("https://%s.gofile.io/uploadFile", strings.TrimSpace(sevData.Data.Server))

	return nil
}

func (b *goFile) PreUpload(_ string, size int64) error {
	if !b.Config.singleMode {
		b.totalSize = size
		b.dataCh = make(chan []byte)
		b.initPipe()
		go b.initMultipartUpload()
		_, _ = b.streamWriter.Write(b.baseBody)
	}
	return nil
}

func (b *goFile) DoUpload(name string, _ int64, file io.Reader) error {

	_, _ = fmt.Fprintf(b.streamWriter, "\r\n--%s\r\n", b.boundary)
	_, _ = fmt.Fprintf(b.streamWriter, "Content-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n", "file", name)
	_, _ = fmt.Fprintf(b.streamWriter, "Content-Type: application/octet-stream\r\n")
	_, _ = fmt.Fprintf(b.streamWriter, "\r\n")
	for {
		buf := make([]byte, 256)
		nr, err := io.ReadFull(file, buf)
		if nr <= 0 {
			break
		}
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Println(err)
			break
		}
		if nr > 0 {
			_, _ = b.streamWriter.Write(buf[:nr])
		}
	}

	return nil
}

func (b *goFile) PostUpload(string, int64) (string, error) {
	if !b.Config.singleMode {
		return b.finishUpload()
	}
	return "", nil
}

func (b *goFile) FinishUpload([]string) (string, error) {
	if b.Config.singleMode {
		return b.finishUpload()
	}
	return "", nil
}

func (b *goFile) finishUpload() (string, error) {
	_, _ = fmt.Fprintf(b.streamWriter, "\r\n--%s--\r\n", b.boundary)
	_ = b.streamWriter.Close()
	sbody, ok := <-b.dataCh
	if !ok {
		return "", fmt.Errorf("internal error, upload failed")
	}
	var sendData respBody
	if err := json.Unmarshal(sbody, &sendData); err != nil {
		return "", fmt.Errorf("parse body returns error: %v", err)
	}
	link := fmt.Sprintf("https://gofile.io/?c=%s", sendData.Data.Code)
	fmt.Printf("Download Link: %s\nRemove Code: %s\n",
		link, sendData.Data.RemovalCode)
	return link, nil
}

func (b *goFile) initPipe() {
	if apis.DebugMode {
		log.Printf("start upload")
	}
	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("category", "file")
	_ = writer.WriteField("comments", "0")

	writerLength := byteBuf.Len()
	writerBody := make([]byte, writerLength)
	_, _ = byteBuf.Read(writerBody)
	_ = writer.Close()

	boundary := writer.Boundary()
	partR, partW := io.Pipe()

	b.baseBody = writerBody
	b.boundary = boundary
	b.streamReader = partR
	b.streamWriter = partW
}

func (b *goFile) initMultipartUpload() {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	//log.Println(b.serverLink, b.streamWriter, b.streamReader)
	req, err := http.NewRequest("POST", b.serverLink, b.streamReader)
	if err != nil {
		close(b.dataCh)
		return
	}
	//req.ContentLength = totalSize
	//req.Header.Set("content-length", strconv.FormatInt(totalSize, 10))
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", b.boundary))
	if apis.DebugMode {
		log.Printf("header: %v", req.Header)
	}
	resp, err := client.Do(req)
	if err != nil {
		if apis.DebugMode {
			log.Printf("do requests returns error: %v", err)
		}
		close(b.dataCh)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if apis.DebugMode {
			log.Printf("read response returns: %v", err)
		}
		close(b.dataCh)
		return
	}
	_ = resp.Body.Close()
	if apis.DebugMode {
		log.Printf("returns: %v", string(body))
	}

	b.dataCh <- body
	return
}

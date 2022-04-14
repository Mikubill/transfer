package methods

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
)

func MultipartUpload(config MultiPartUploadConfig) ([]byte, error) {
	if config.Debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_, err := writer.CreateFormFile("file", config.FileName)
	if err != nil {
		return nil, err
	}

	writerLength := byteBuf.Len()
	writerBody := make([]byte, writerLength)
	_, _ = byteBuf.Read(writerBody)
	_ = writer.Close()

	lastBoundary := fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary())
	totalSize := int64(writerLength) + config.FileSize + int64(len(lastBoundary))
	partR, partW := io.Pipe()

	go func() {
		_, _ = partW.Write(writerBody)
		for {
			buf := make([]byte, 256)
			nr, err := io.ReadFull(config.FileReader, buf)
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

	req, err := http.NewRequest("POST", config.Endpoint, partR)
	if err != nil {
		return nil, err
	}
	req.ContentLength = totalSize
	req.Header.Set("content-length", strconv.FormatInt(totalSize, 10))
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	if config.Debug {
		log.Printf("header: %v", req.Header)
	}
	resp, err := client.Do(req)
	if err != nil {
		if config.Debug {
			log.Printf("do requests returns error: %v", err)
		}
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.Debug {
			log.Printf("read response returns: %v", err)
		}
		return nil, err
	}
	_ = resp.Body.Close()
	if config.Debug {
		log.Printf("returns: %v", string(body))
	}

	return body, nil
}

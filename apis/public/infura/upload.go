package infura

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
)

const upload = "https://ipfs.infura.io:5001/api/v0/add?recursive=false"

type InfResp struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

func (b *infura) DoUpload(name string, size int64, file io.Reader) error {

	body, err := b.newMultipartUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      apis.DebugMode,
	})
	if err != nil {
		return fmt.Errorf("upload returns error: %s", err)
	}

	var s InfResp
	if err := json.Unmarshal(body, &s); err != nil {
		return fmt.Errorf("unmarshal returns error: %s", err)
	}

	b.resp = s.Hash
	return nil
}

func (b *infura) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: https://ipfs.io/ipfs/%s\n", b.resp)
	return b.resp, nil
}

func (b infura) newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_, err := writer.CreateFormFile("file", config.fileName)
	if err != nil {
		return nil, err
	}

	writerLength := byteBuf.Len()
	writerBody := make([]byte, writerLength)
	_, _ = byteBuf.Read(writerBody)
	_ = writer.Close()

	lastBoundary := []byte(fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary()))
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
		partW.Write(lastBoundary)
		_ = partW.Close()
	}()

	req, err := http.NewRequest("POST", upload, partR)
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

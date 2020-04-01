package gofile

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"transfer/utils"
)

const (
	getServer = "https://apiv2.gofile.io/getServer"
)

func (b goFile) Upload(files []string) {
	for _, v := range files {
		b.initUpload([]string{v})
	}
}

func (b goFile) initUpload(files []string) {

	for _, v := range files {
		if utils.IsExist(v) {
			err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				err = b.upload(path)
				if err != nil {
					fmt.Printf("upload returns error: %v, onfile: %s\n", err, path)
				}
				return nil
			})
			if err != nil {
				fmt.Printf("filepath.walk(upload) returns error: %v, onfile: %s\n", err, v)
			}
		} else {
			fmt.Printf("%s not found\n", v)
		}
	}
}

func (b goFile) upload(v string) error {
	fmt.Printf("Local: %s\n", v)
	if b.Config.debugMode {
		log.Println("retrieving file info...")
	}
	info, err := utils.GetFileInfo(v)
	if err != nil {
		return fmt.Errorf("getFileInfo returns error: %v", err)
	}

	file, err := os.Open(v)
	if err != nil {
		return fmt.Errorf("openFile returns error: %v", err)
	}

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
	serverLink := fmt.Sprintf("https://%s.gofile.io/upload", strings.TrimSpace(sevData.Data.Server))

	bar := pb.Full.Start64(info.Size())
	sbody, err := b.newMultipartUpload(serverLink, uploadConfig{
		fileSize:   info.Size(),
		fileName:   info.Name(),
		fileReader: bar.NewProxyReader(file),
		debug:      b.Config.debugMode,
	})
	if err != nil {
		return fmt.Errorf("post %s returns error: %s", serverLink, err)
	}

	var sendData respBody
	if err := json.Unmarshal(sbody, &sendData); err != nil {
		return fmt.Errorf("parse body returns error: %v", err)
	}

	_ = file.Close()
	bar.Finish()
	fmt.Printf("Download Link: https://gofile.io/?c=%s\n", sendData.Data.Code)
	fmt.Printf("Direct Link: https://%s.gofile.io/download/%s/%s\n",
		strings.TrimSpace(sevData.Data.Server), sendData.Data.Code, info.Name())
	fmt.Printf("Remove Code: %s\n", sendData.Data.RemovalCode)
	return nil
}

func (b goFile) newMultipartUpload(link string, config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	// todo: fix cert error
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("category", "file")
	_ = writer.WriteField("comments", "0")
	_, err := writer.CreateFormFile("filesUploaded", config.fileName)
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
		buf := make([]byte, 256)
		for {
			nr, err := config.fileReader.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("error reading from connector: %v", err)
			}
			if nr > 0 {
				_, _ = partW.Write(buf[:nr])
			}
		}
		_, _ = partW.Write(lastBoundary)
		_ = partW.Close()
	}()

	req, err := http.NewRequest("POST", link, partR)
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

package bitsend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"transfer/utils"
)

const download = "https://bitsend.jp/download/%s.html"

func (b bitSend) Upload(files []string) {
	for _, v := range files {
		b.initUpload([]string{v})
	}
}

func (b bitSend) initUpload(files []string) {

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

func (b bitSend) upload(v string) error {
	fmt.Printf("Local: %s\n", v)
	if b.Config.debugMode {
		log.Println("retrieving file info...")
	}
	info, err := utils.GetFileInfo(v)
	if err != nil {
		return fmt.Errorf("getFileInfo returns error: %v", err)
	}

	bar := pb.Full.Start64(info.Size())
	file, err := os.Open(v)
	if err != nil {
		return fmt.Errorf("openFile returns error: %v", err)
	}

	body, err := newMultipartUpload(uploadConfig{
		fileSize:   info.Size(),
		fileName:   path.Base(v),
		fileReader: bar.NewProxyReader(file),
		debug:      b.Config.debugMode,
	})

	respDat := new(uploadResp)
	err = json.Unmarshal(body, respDat)
	if err != nil {
		return fmt.Errorf("post %s returns error: ", err)
	}
	if len(respDat.Files) == 0 {
		return fmt.Errorf("upload %s failed: no file found ", v)
	}
	if b.Config.debugMode {
		log.Printf("%+v", respDat)
	}

	_ = file.Close()
	bar.Finish()

	fmt.Println("Download Link: " + fmt.Sprintf(download, respDat.Files[0].FileKey))
	fmt.Println("Delete Link: " + respDat.Files[0].DeleteUrl)

	return nil
}

func newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("u_key", utils.GenRandString(16))
	_, err := writer.CreateFormFile("files[]", config.fileName)
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

	req, err := http.NewRequest("POST", "https://bitsend.jp/jqu/", partR)
	if err != nil {
		return nil, err
	}
	req.ContentLength = totalSize
	req.Header.Set("referer", "https://bitsend.jp/")
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

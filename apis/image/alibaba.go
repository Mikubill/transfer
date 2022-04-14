package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Mikubill/transfer/utils"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"time"
)

var (
	AliBackend = new(Ali)
)

type Ali struct {
	picBed
}

type AliResp struct {
	Code string `json:"code"`
	Url  string `json:"url"`
	Hash string `json:"hash"`
}

// func (s Ali) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("H[0-9a-zA-Z]+")
// 	return matcher.FindString(link)
// }

func (s Ali) linkBuilder(link string) string {
	getter := regexp.MustCompile("H[0-9a-zA-Z]+")
	return "https://ae01.alicdn.com/kf/" + getter.FindString(link) + ".jpg"
}

// func (s Ali) UploadStream(dataChan chan UploadDataFlow) {
// 	for {
// 		data, ok := <-dataChan
// 		if !ok {
// 			break
// 		}
// 		url, err := s.Upload(data.Data)
// 		if err != nil {
// 			dataChan <- data
// 			continue
// 		}
// 		data.HashMap.Set(strconv.FormatInt(data.Offset, 10), s.linkExtractor(url))
// 		data.Wg.Done()
// 	}
// }

// func (s Ali) DownloadStream(dataChan chan DownloadDataFlow) {
// 	for {
// 		data, ok := <-dataChan
// 		if !ok {
// 			break
// 		}
// 		link := s.linkBuilder(data.Hash)
// 		resp, err := http.Get(link)
// 		if err != nil {
// 			dataChan <- data
// 			continue
// 		}
// 		bd, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			dataChan <- data
// 			continue
// 		}
// 		_ = resp.Body.Close()
// 		offset, _ := strconv.ParseInt(data.Offset, 10, 64)
// 		n, _ := data.File.WriteAt(bd, offset)
// 		data.Bar.Add(n)
// 		data.Wg.Done()
// 	}
// }

func (s Ali) Upload(data []byte) (string, error) {
	client := http.Client{Timeout: 30 * time.Second}
	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	filename := utils.GenRandString(14) + ".jpg"
	_ = writer.WriteField("name", filename)
	_ = writer.WriteField("scene", "aeMessageCenterV2ImageRule")
	w, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	_, _ = w.Write(data)
	_ = writer.Close()
	req, err := http.NewRequest("POST", "https://kfupload.alibaba.com/mupload", byteBuf)
	if err != nil {
		return "", err
	}

	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var r AliResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	if r.Code != "0" {
		return "", fmt.Errorf(string(body))
	}
	return s.linkBuilder(r.Url), nil
}

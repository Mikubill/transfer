package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"time"
	"transfer/utils"
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

func (s Ali) linkExtractor(link string) string {
	matcher := regexp.MustCompile("H[0-9a-zA-Z]+")
	return matcher.FindString(link)
}

func (s Ali) linkBuilder(link string) string {
	getter := regexp.MustCompile("H[0-9a-zA-Z]+")
	return "https://ae01.alicdn.com/kf/" + getter.FindString(link) + ".jpg"
}

func (s Ali) Upload(data []byte) (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
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
	_ = resp.Body.Close()

	var r AliResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	return s.linkBuilder(r.Url), nil
}

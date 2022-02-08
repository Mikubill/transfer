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
	BDBackend = new(BD)
)

type BD struct {
	picBed
}

type BDResp struct {
	Message string `json:"msg"`
	Data    BDItem `json:"data"`
}

type BDItem struct {
	Sign string `json:"sign"`
}

func (s BD) linkExtractor(link string) string {
	matcher := regexp.MustCompile("[0-9a-f]{32}")
	return matcher.FindString(link)
}

func (s BD) linkBuilder(link string) string {
	getter := regexp.MustCompile("[0-9a-f]{32}")
	return "https://graph.baidu.com/resource/" + getter.FindString(link) + ".jpg"
}

func (s BD) Upload(data []byte) (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	filename := utils.GenRandString(14) + ".jpg"
	_ = writer.WriteField("pos", "upload")
	_ = writer.WriteField("uptype", "upload_pc")
	_ = writer.WriteField("fm", "index")
	w, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return "", err
	}
	_, _ = w.Write(data)
	_ = writer.Close()
	req, err := http.NewRequest("POST", "https://graph.baidu.com/upload", byteBuf)
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

	var r BDResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}

	if r.Message != "Success" {
		return "", fmt.Errorf(r.Message)
	}

	return s.linkBuilder(r.Data.Sign), nil
}

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
	SNBackend = new(SN)
)

type SN struct {
	picBed
}

type SNResp struct {
	Url string `json:"src"`
}

// func (s SN) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[0-9]{18}")
// 	return matcher.FindString(link)
// }

func (s SN) linkBuilder(link string) string {
	getter := regexp.MustCompile("[0-9]{18}")
	return "https://image.suning.cn/uimg/ZR/share_order/" + getter.FindString(link) + ".jpg"
}

func (s SN) Upload(data []byte) (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	filename := utils.GenRandString(14) + ".jpg"
	_ = writer.WriteField("omsOrderItemId", "1")
	_ = writer.WriteField("custNum", "1")
	_ = writer.WriteField("deviceType", "1")
	w, err := writer.CreateFormFile("Filedata", filename)
	if err != nil {
		return "", err
	}
	_, _ = w.Write(data)
	_ = writer.Close()
	req, err := http.NewRequest("POST", "http://review.suning.com/imageload/uploadImg.do", byteBuf)
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

	var r SNResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	return s.linkBuilder(r.Url), nil
}

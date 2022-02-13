package image

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	TTBackend = new(TT)
)

type TT struct {
	picBed
}

type TTResp struct {
	Code string `json:"message"`
	D    string `json:"web_url"`
}

// func (s TT) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[a-zA-Z0-9]{20}")
// 	return matcher.FindString(link)
// }

func (s TT) linkBuilder(link string) string {
	getter := regexp.MustCompile("[a-zA-Z0-9]{20}")
	return "https://p.pstatp.com/origin/" + getter.FindString(link)
}

func (s TT) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://mp.toutiao.com/upload_photo/?type=json", "photo", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r TTResp

	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}

	if r.Code != "success" {
		return "", fmt.Errorf(r.Code)
	}

	return s.linkBuilder(r.D), nil
}

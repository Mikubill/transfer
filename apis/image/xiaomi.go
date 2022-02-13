package image

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	XMBackend = new(XM)
)

type XM struct {
	picBed
}

type XMResp struct {
	Code string `json:"message"`
	D    string `json:"result"`
}

// func (s XM) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[a-f0-9]{32}")
// 	return matcher.FindString(link)
// }

func (s XM) linkBuilder(link string) string {
	geXMer := regexp.MustCompile("[a-f0-9]{32}")
	return "https://shop.io.mi-img.com/app/shop/img?id=shop_" + geXMer.FindString(link) + ".png"
}

func (s XM) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://shopapi.io.mi.com/homemanage/shop/uploadpic", "pic", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r XMResp

	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}

	if r.Code != "ok" {
		return "", fmt.Errorf(r.Code)
	}

	return s.linkBuilder(r.D), nil
}

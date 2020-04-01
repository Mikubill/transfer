package image

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	CCBackend = new(CC)
)

type CC struct {
	picBed
}

type CCResp struct {
	Code  string        `json:"code"`
	Image []CCImageItem `json:"image"`
}

type CCImageItem struct {
	URL string `json:"url"`
	Del string `json:"delete"`
}

func (s CC) linkExtractor(link string) string {
	matcher := regexp.MustCompile("i[0-9]/[0-9]{4}/[0-9]{2}/[0-9]{2}/[a-zA-Z0-9]+")
	return matcher.FindString(link)
}

func (s CC) linkBuilder(link string) string {
	getter := regexp.MustCompile("i[0-9]/[0-9]{4}/[0-9]{2}/[0-9]{2}/[a-zA-Z0-9]+")
	return "https://upload.cc/" + getter.FindString(link) + ".png"
}

func (s CC) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://upload.cc/image_upload", "file")
	if err != nil {
		return "", err
	}

	var r CCResp
	if strings.Contains(string(body), "error") {
		return "", fmt.Errorf(string(body))
	}

	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	return s.linkBuilder(r.Image[0].URL), nil
}

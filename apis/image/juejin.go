package image

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	JJBackend = new(JJ)
)

type JJ struct {
	picBed
}

type JJResp struct {
	Code int   `json:"s"`
	D    JJCol `json:"d"`
}

type JJCol struct {
	URL JJItem `json:"url"`
}

type JJItem struct {
	Http  string `json:"http"`
	Https string `json:"https"`
}

// func (s JJ) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[0-9]{4}/[0-9]{1,2}/[0-9]{1,2}/[a-zA-Z0-9]+")
// 	return matcher.FindString(link)
// }

func (s JJ) linkBuilder(link string) string {
	getter := regexp.MustCompile("[0-9]{4}/[0-9]{1,2}/[0-9]{1,2}/[a-zA-Z0-9]+")
	return "https://user-gold-cdn.xitu.io/" + getter.FindString(link)
}

func (s JJ) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://cdn-ms.juejin.im/v1/upload?bucket=gold-user-assets", "file", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r JJResp
	if strings.Contains(string(body), "error") {
		return "", fmt.Errorf(string(body))
	}

	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	return s.linkBuilder(r.D.URL.Https), nil
}

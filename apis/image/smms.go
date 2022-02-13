package image

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	SMBackend = new(SM)
)

type SM struct {
	picBed
}

type SMResp struct {
	Code  string `json:"code"`
	Error string `json:"message"`
	Data  SMItem `json:"data"`
}

type SMItem struct {
	URL string `json:"url"`
}

// func (s SM) linkExtractor(link string) string {
// 	link = strings.ReplaceAll(link, "\\/", "/")
// 	matcher := regexp.MustCompile("[0-9]{4}/[0-9]{2}/[0-9]{2}/[a-zA-Z0-9]{15}")
// 	return matcher.FindString(link)
// }

func (s SM) linkBuilder(link string) string {
	link = strings.ReplaceAll(link, "\\/", "/")
	getter := regexp.MustCompile("[0-9]{4}/[0-9]{2}/[0-9]{2}/[a-zA-Z0-9]{15}")
	return "http://i.loli.net/" + getter.FindString(link) + ".png"
}

func (s SM) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://sm.ms/api/v2/upload", "smfile", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r SMResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	if r.Code != "success" {
		if strings.Contains(r.Error, "exists at") {
			return s.linkBuilder(r.Error), nil
		}
		return "", fmt.Errorf(r.Error)
	}
	return s.linkBuilder(r.Data.URL), nil
}

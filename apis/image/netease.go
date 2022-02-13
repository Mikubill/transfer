package image

import (
	"encoding/json"
	"regexp"
)

var (
	NTBackend = new(NT)
)

type NT struct {
	picBed
}

type NTResp struct {
	Code  string   `json:"code"`
	Error string   `json:"errorCode"`
	Data  []string `json:"data"`
}

// func (s NT) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[a-zA-Z0-9]{32}")
// 	return matcher.FindString(link)
// }

func (s NT) linkBuilder(link string) string {
	getter := regexp.MustCompile("[a-zA-Z0-9]{32}")
	return "http://yanxuan.nosdn.127.net/" + getter.FindString(link) + ".png"
}

func (s NT) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "http://you.163.com/xhr/file/upload.json", "file", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r NTResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	if r.Code != "200" {
		return r.Error, nil
	}
	return s.linkBuilder(r.Data[0]), nil
}

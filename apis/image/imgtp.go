package image

import (
	"encoding/json"
)

var (
	ItpBackend = new(Itp)
)

type Itp struct {
	picBed
}

type ItpResp struct {
	Code int64 `json:"code"`
	Data struct {
		ID   string `json:"id"`
		Md5  string `json:"md5"`
		Mime string `json:"mime"`
		Name string `json:"name"`
		Sha1 string `json:"sha1"`
		Size int64  `json:"size"`
		URL  string `json:"url"`
	} `json:"data"`
	Msg  string `json:"msg"`
	Time int64  `json:"time"`
}

// func (s TG) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[a-f0-9]+\\.\\w+")
// 	return matcher.FindString(link)
// }

func (s Itp) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://imgtp.com/upload/upload", "image", defaultReqMod)
	if err != nil {
		return "", err
	}

	var p ItpResp

	if json.Unmarshal(body, &p); err != nil {
		return "", err
	}

	return p.Data.URL, nil
}

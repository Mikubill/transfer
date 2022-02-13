package image

import (
	"encoding/json"
)

var (
	IKrBackend = new(IKr)
)

type IKr struct {
	picBed
}

type IKrResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    string `json:"data"`
}

// func (s IKr) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("https://s3.bmp.ovh/")
// 	return matcher.FindString(link)
// }

func (s IKr) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://imgkr.com/api/v2/files/upload", "file", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r IKrResp
	if json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	return r.Data, nil
}

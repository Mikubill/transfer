package image

import (
	"encoding/json"
)

var (
	IUBackend = new(IU)
)

type IU struct {
	picBed
}

type IUResp struct {
	Code         int    `json:"code"`
	ID           int    `json:"id"`
	Imgid        string `json:"imgid"`
	RelativePath string `json:"relative_path"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Delete       string `json:"delete"`
}

// func (s IU) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("https://s3.bmp.ovh/")
// 	return matcher.FindString(link)
// }

func (s IU) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://imgurl.org/upload/aws_s3", "file", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r IUResp
	if json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	return r.URL, nil
}

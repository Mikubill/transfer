package image

import (
	"regexp"
	"strings"
)

var (
	TGBackend = new(TG)
)

type TG struct {
	picBed
}

// func (s TG) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[a-f0-9]+\\.\\w+")
// 	return matcher.FindString(link)
// }

func (s TG) linkBuilder(link string) string {
	geVMer := regexp.MustCompile(`[a-f0-9]+\.\w+`)
	return "https://telegra.ph/file/" + geVMer.FindString(link)
}

func (s TG) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://telegra.ph/upload", "file", defaultReqMod)
	if err != nil || strings.Contains(string(body), "error") {
		return "", err
	}

	return s.linkBuilder(string(body)), nil
}

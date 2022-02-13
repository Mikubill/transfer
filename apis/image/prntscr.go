package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var (
	PRBackend = new(PR)
	PRFinder  = regexp.MustCompile(`og:image" content="(.*?)"`)
)

type PR struct {
	picBed
}

type PRResp struct {
	Code string `json:"status"`
	Data string `json:"data"`
}

// func (s PR) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[a-zA-Z0-9]{22}")
// 	return matcher.FindString(link)
// }

// func (s PR) linkBuilder(link string) string {
// 	getter := regexp.MustCompile("[a-zA-Z0-9]{22}")
// 	return "https://image.prntscr.com/image/" + getter.FindString(link) + ".png"
// }

func (s PR) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://prntscr.com/upload.php", "image", defaultReqMod)
	if err != nil {
		return "", err
	}

	var r PRResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	if r.Code != "success" {
		return r.Data, nil
	}

	link := strings.ReplaceAll(r.Data, "\\/", "/")
	if Verbose {
		fmt.Println("requesting: " + link)
	}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:65.0) Gecko/20100101 Firefox/65.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	return string(PRFinder.FindSubmatch(body)[1]), nil
}

package lanzous

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/Mikubill/transfer/apis"
)

var (
	matcher = regexp.MustCompile("(https://)?www\\.lanzous\\.com/[0-9a-z]+")
	regex   = regexp.MustCompile("/fn\\?[A-Za-z_0-9]+_c")
	regex2  = regexp.MustCompile("[A-Za-z_0-9]+_c")
)

type vsResp struct {
	Domain string `json:"dom"`
	Key    string `json:"url"`
}

func (b lanzous) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}

func (b lanzous) DoDownload(link string, config apis.DownConfig) error {
	body, err := httpRequest("GET", link, "")
	if err != nil {
		return err
	}
	c1 := string(regex.Find(body))

	if c1 == "" {
		return fmt.Errorf("failed getting link")
	}

	linkR := "https://www.lanzous.com/" + c1
	if apis.DebugMode {
		log.Println(linkR)
	}
	body, err = httpRequest("GET", linkR, "")
	if err != nil {
		return err
	}
	c1 = string(regex2.Find(body))

	if c1 == "" {
		return fmt.Errorf("failed getting link" + string(body))
	}

	pb := fmt.Sprintf("action=downprocess&sign=%s&ves=1", c1)
	if apis.DebugMode {
		log.Println(pb)
	}
	body, err = httpRequest("POST", "https://www.lanzous.com/ajaxm.php", pb)
	if err != nil {
		return err
	}
	//log.Println(string(body))
	var vRe vsResp
	if err := json.Unmarshal(body, &vRe); err != nil {
		return err
	}

	finLink := vRe.Domain + "/file/?" + vRe.Key
	if apis.DebugMode {
		log.Println(finLink)
	}
	err = b.download(finLink, config)
	if err != nil {
		return fmt.Errorf("download failed on %s, returns %s\n", link, err)
	}
	return nil
}

func (b lanzous) download(v string, config apis.DownConfig) error {
	config.Parallel = 1 // force
	config.Link = v
	config.Modifier = apis.AddHeaders

	err := apis.DownloadFile(config)
	if err != nil {
		return err
	}
	return nil
}

func httpRequest(action string, link string, content string) ([]byte, error) {
	req, err := http.NewRequest(action, link, strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "accept: application/json, text/javascript, */*")
	req.Header.Add("user-agent", "Mozilla/5.0")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;")
	req.Header.Add("referer", "https://www.lanzous.com/")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	return body, nil
}

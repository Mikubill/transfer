package musetransfer

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

var (
	Regex   = regexp.MustCompile(`https?://musetransfer\.com/s/([a-z0-9]{6,})`)
	dlAPI   = "https://service.tezign.com/transfer/share/download"
	listAPI = "https://service.tezign.com/transfer/share/file/list"
)

func (b muse) LinkMatcher(v string) bool {
	return Regex.MatchString(v)
}

func (b muse) DoDownload(link string, config apis.DownConfig) error {
	err := b.download(link, config)
	if err != nil {
		fmt.Printf("download failed on %s, returns %s\n", link, err)
	}
	return nil
}

func (b muse) download(v string, config apis.DownConfig) error {
	fmt.Printf("fetching ticket..")
	b.Config.devicetoken = utils.GenRandString(8)[:11]
	end := utils.DotTicker()

	subgroup := Regex.FindStringSubmatch(v)
	if len(subgroup) != 2 {
		return fmt.Errorf("invalid muse link: %s", v)
	}
	qform := map[string]any{
		"code": subgroup[1],
	}

	// request filelist
	buf, _ := json.Marshal(qform)
	body, err := b.postAPI(listAPI, buf)
	if err != nil {
		return fmt.Errorf("request %s error: %v", listAPI, err)
	}

	var listResp fsListResp
	err = json.Unmarshal(body, &listResp)
	if err != nil {
		return err
	}
	if listResp.Message != "success" {
		return fmt.Errorf("get file list error: %s", listResp.Message)
	}
	if len(listResp.Result) == 0 {
		return fmt.Errorf("no file found in url: %s", v)
	}
	*end <- struct{}{}
	fmt.Printf("ok\n")

	for _, item := range listResp.Result {
		qform["fileId"] = item.FileID
		buf, _ = json.Marshal(qform)
		body, err := b.postAPI(dlAPI, buf)
		if err != nil {
			return err
		}

		// decode to dlresp
		var dlresp dlResp
		err = json.Unmarshal(body, &dlresp)
		if err != nil {
			return err
		}

		if dlresp.Message != "success" {
			return fmt.Errorf("download failed on %s, returns %s", item.Name, dlresp.Message)
		}

		config.Link = dlresp.Result.URL
		config.Modifier = addHeaders

		err = apis.DownloadFile(config)
		if err != nil {
			return err
		}
	}
	return nil
}

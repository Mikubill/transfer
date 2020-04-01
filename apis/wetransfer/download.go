package wetransfer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
	"transfer/utils"
)

var (
	signDownload string
	regexShorten = regexp.MustCompile("https://we\\.tl/[a-zA-Z0-9\\-]{12}")
	regex        = regexp.MustCompile("https://wetransfer\\.com/downloads/[a-z0-9]{46}/[a-z0-9]{6}")
	regexList    = regexp.MustCompile("{\"id.*}]}")
)

func (b weTransfer) Download(files []string) {
	for _, v := range files {
		err := b.download(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b weTransfer) download(v string) error {
	client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
	fmt.Printf("fetching ticket..")
	end := utils.DotTicker()

	if !regexShorten.MatchString(v) && !regex.MatchString(v) {
		return fmt.Errorf("url is invaild")
	}
	if b.Config.debugMode {
		log.Println("step 1/2 metadata")
		log.Printf("link: %+v", v)
	}
	resp, err := client.Get(v)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	tk := tokenRegex.FindSubmatch(body)
	if len(tk) == 0 {
		return fmt.Errorf("no csrf-token found")
	}
	ticket := requestTicket{
		token:   string(tk[1]),
		cookies: "",
	}
	ck := resp.Header.Values("Set-Cookie")
	for _, v := range ck {
		s := strings.Split(v, ";")
		ticket.cookies += s[0] + ";"
	}
	if b.Config.debugMode {
		log.Printf("ticket: %+v", ticket)
	}
	_ = resp.Body.Close()
	dat := regexList.FindString(string(body))
	if b.Config.debugMode {
		log.Printf("dst: %+v", dat)
	}
	var block configBlock
	if err := json.Unmarshal([]byte(dat), &block); err != nil {
		return err
	}
	signDownload = fmt.Sprintf("https://wetransfer.com/api/v4/transfers/%s/download", block.ID)

	*end <- struct{}{}
	fmt.Printf("ok\n")
	for _, item := range block.Item {
		err = b.downloadItem(item, block.Hash, ticket)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (b weTransfer) downloadItem(item fileInfo, token string, tk requestTicket) error {
	if b.Config.debugMode {
		log.Println("step2 -> api/getConf")
	}
	data, _ := json.Marshal(map[string]interface{}{
		"security_hash":  token,
		"domain_user_id": utils.GenRandUUID(),
		"file_ids":       []string{item.ID},
	})
	if b.Config.debugMode {
		log.Printf("tk: %+v", tk)
	}
	resp, err := newRequest(signDownload, string(data), requestConfig{
		action:   "POST",
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(tk),
	})
	if err != nil {
		return fmt.Errorf("sign Request returns error: %s, onfile: %s", err, item.Name)
	}

	if b.Config.debugMode {
		log.Println("step3 -> startDownload")
	}
	filePath := b.Config.prefix

	if utils.IsExist(b.Config.prefix) {
		if utils.IsFile(b.Config.prefix) {
			filePath = b.Config.prefix
		} else {
			filePath = path.Join(b.Config.prefix, item.Name)
		}
	}

	//fmt.Printf("File save to: %s\n", filePath)
	err = utils.DownloadFile(filePath, resp.Download, utils.DownloadConfig{
		Force:    b.Config.forceMode,
		Debug:    b.Config.debugMode,
		Parallel: b.Config.parallel,
		Modifier: addHeaders,
	})
	if err != nil {
		return fmt.Errorf("failed DownloadConfig with error: %s, onfile: %s", err, item.Name)
	}
	return nil
}

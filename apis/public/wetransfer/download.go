package wetransfer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

var (
	signDownload, safetyHash, blockID string
	regexShorten                      = regexp.MustCompile("(https://)?we\\.tl/[a-zA-Z0-9\\-]{12}")
	regex                             = regexp.MustCompile("https?://wetransfer\\.com/downloads/([a-z0-9]{46})/([a-z0-9]{6})")
)

func (b weTransfer) LinkMatcher(v string) bool {
	return regex.MatchString(v) || regexShorten.MatchString(v)
}

func (b weTransfer) DoDownload(link string, config apis.DownConfig) error {
	err := b.download(link, config)
	if err != nil {
		fmt.Printf("download failed on %s, returns %s\n", link, err)
	}
	return nil
}

func (b weTransfer) download(v string, config apis.DownConfig) error {
	client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
	fmt.Printf("fetching ticket..")
	end := utils.DotTicker()

	if regexShorten.MatchString(v) {

		req, err := http.NewRequest("HEAD", v, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultTransport.RoundTrip(req)
		v = resp.Header.Get("Location")
	}

	tk0 := regex.FindStringSubmatch(v)
	if len(tk0) < 3 || !regex.MatchString(v) {
		return fmt.Errorf("url is invaild")
	}
	blockID = string(tk0[1])
	safetyHash = string(tk0[2])

	if apis.DebugMode {
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
	if apis.DebugMode {
		log.Printf("ticket: %+v", ticket)
	}
	_ = resp.Body.Close()

	signPreDownload := fmt.Sprintf("https://wetransfer.com/api/v4/transfers/%s/prepare-download", blockID)
	data, _ := json.Marshal(map[string]any{
		"security_hash": safetyHash,
	})
	if apis.DebugMode {
		log.Printf("tk: %+v", tk)
	}
	resp0, err := newRequest(signPreDownload, string(data), requestConfig{
		action:   "POST",
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})

	// var block configBlock
	// if err := json.Unmarshal([]byte(resp0), &block); err != nil {
	// 	return err
	// }
	signDownload = fmt.Sprintf("https://wetransfer.com/api/v4/transfers/%s/download", resp0.ID)

	*end <- struct{}{}
	fmt.Printf("ok\n")
	if resp0.State != "downloadable" {
		return fmt.Errorf("link state is not downloadable (state: %s)", resp0.State)
	}

	for _, item := range resp0.Item {
		config.Ticket = resp0.Hash
		err = b.downloadItem(item, ticket, config)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (b weTransfer) downloadItem(item fileInfo, tk requestTicket, config apis.DownConfig) error {
	if apis.DebugMode {
		log.Println("step2 -> api/getConf")
	}
	data, _ := json.Marshal(map[string]any{
		"security_hash":  config.Ticket,
		"domain_user_id": utils.GenRandUUID(),
		"file_ids":       []string{item.ID},
		"intent":         "single_file",
	})
	if apis.DebugMode {
		log.Printf("tk: %+v", tk)
	}
	resp, err := newRequest(signDownload, string(data), requestConfig{
		action:   "POST",
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(tk),
	})
	if err != nil {
		return fmt.Errorf("sign Request error: %s, onfile: %s", err, item.Name)
	}

	if apis.DebugMode {
		log.Println("step3 -> startDownload")
	}
	filePath, err := filepath.Abs(config.Prefix)
	if err != nil {
		return fmt.Errorf("parse filepath error: %s, onfile: %s", err, item.Name)
	}

	if utils.IsExist(filePath) && utils.IsDir(filePath) {
		filePath = path.Join(filePath, item.Name)
	}

	config.Prefix = filePath
	config.Link = resp.Download
	config.Modifier = addHeaders

	err = apis.DownloadFile(config)
	if err != nil {
		return fmt.Errorf("download failed: %s, onfile: %s", err, item.Name)
	}
	return nil
}

package bitsend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

var (
	regex   = regexp.MustCompile("files/[0-9a-z]{32}\\.")
	matcher = regexp.MustCompile("(https://)?bitsend\\.jp/download/[0-9a-z]{32}(\\.html)?")
)

func (b bitSend) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}

func (b bitSend) DoDownload(link string, config apis.DownConfig) error {
	return b.download(link, config)
}

func (b bitSend) download(v string, config apis.DownConfig) error {
	var link string
	if apis.DebugMode {
		log.Printf("fetching: %v", v)
	}
	fmt.Printf("fetching ticket..")
	end := utils.DotTicker()

	resp, err := http.Get(v)
	if err != nil {
		fmt.Printf("failed fetching %s: %v", v, err)
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read response error: %v", err)
		return nil
	}
	_ = resp.Body.Close()
	link = "https://bitsend.jp/" + regex.FindString(string(body))
	if apis.DebugMode {
		log.Printf("dest: %v", link)
	}
	b.Ticket = resp.Header.Get("Set-Cookie")
	*end <- struct{}{}
	fmt.Printf("ok\n")

	config.Link = link
	config.Modifier = b.addRef(v)

	return apis.DownloadFile(config)
}

func (b bitSend) addRef(ref string) func(req *http.Request) {
	return func(req *http.Request) {
		addHeaders(req)

		req.Header.Set("Referer", ref)
		req.Header.Set("cookie", b.headRequest(ref, 0))
		//log.Printf("%+v", req.Header)
	}
}

func (b bitSend) headRequest(link string, retry int) string {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(link)
	if err != nil {
		if retry > 2 {
			panic("Failed to get download token.")
		}
		retry++
		return b.headRequest(link, retry)
	}
	b.Ticket = resp.Header.Get("Set-Cookie")
	return b.Ticket
}

func addHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Chrome/80.0.3987.149 bitsend-Uploader")
	req.Header.Set("Origin", "https://bitsend.jp/")
}

package bitsend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"transfer/utils"
)

var (
	regex = regexp.MustCompile("files/[0-9a-z]{32}\\.")
)

func (b bitSend) Download(files []string) {
	for _, v := range files {
		err := b.download(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b bitSend) download(v string) error {
	var link string
	if strings.Contains(v, "/download") {
		if b.Config.debugMode {
			log.Printf("fetching: %v", v)
		}
		fmt.Printf("fetching download metadata..")
		end := utils.DotTicker()

		resp, err := http.Get(v)
		if err != nil {
			return fmt.Errorf("failed fetching %s: %v", v, err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if b.Config.debugMode {
				log.Printf("read response returns: %v", err)
			}
			return err
		}
		_ = resp.Body.Close()
		link = "https://bitsend.jp/" + regex.FindString(string(body))
		if b.Config.debugMode {
			log.Printf("dest: %v", link)
		}
		b.Ticket = resp.Header.Get("Set-Cookie")
		*end <- struct{}{}
		fmt.Printf("ok\n")
	} else {
		return fmt.Errorf("url format invalid: %v", v)
	}

	err := utils.DownloadFile(b.Config.prefix, link, utils.DownloadConfig{
		Force:    b.Config.forceMode,
		Debug:    b.Config.debugMode,
		Parallel: b.Config.parallel,
		Modifier: b.addRef(v),
	})
	if err != nil {
		return fmt.Errorf("failed fetching %s: %v", link, err)
	}
	return nil
}

func (b bitSend) addRef(ref string) func(req *http.Request) {
	return func(req *http.Request) {
		addHeaders(req)

		req.Header.Set("Referer", ref)
		req.Header.Set("cookie", b.headRequest(ref, 0))
		log.Printf("%+v", req.Header)
	}
}

func (b bitSend) headRequest(link string, retry int) string {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(link)
	if err != nil {
		if b.Config.debugMode {
			log.Printf("error heading %s: %v\n", link, err)
		}
		if retry > 2 {
			return b.Ticket
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

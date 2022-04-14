package gofile

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

var (
	matcher = regexp.MustCompile("(https://)?gofile\\.io/(\\?c=|download/|d/)[0-9a-zA-Z]{6}")
	reg     = regexp.MustCompile("=[0-9a-zA-Z]{6}")
)

func (b goFile) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}

func (b goFile) DoDownload(link string, config apis.DownConfig) error {
	err := b.download(link, config)
	if err != nil {
		return fmt.Errorf("download failed on %s, returns %s", link, err)
	}
	return nil
}

func (b goFile) download(v string, config apis.DownConfig) error {
	fileID := reg.FindString(v)
	fileID = fileID[1:]

	var sevData folderDetails2
	fmt.Printf("fetching ticket..")
	end := utils.DotTicker()

	err := b.createUser()
	if err != nil {
		return err
	}

	body, err := http.Get(fmt.Sprintf("https://api.gofile.io/getContent?contentId=%s&token=%s", fileID, b.userToken))
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}

	err = smallParser(body, &sevData)
	if err != nil {
		return fmt.Errorf("request %s error: %v", getServer, err)
	}
	*end <- struct{}{}
	if apis.DebugMode {
		log.Printf("retuens: %+v", sevData)
	}
	fmt.Printf("done\n")

	for _, item := range sevData.Data.Contents {
		if apis.DebugMode {
			log.Printf("fileitem: %+v\n", item)
		}

		config.Link = item.Link
		config.Modifier = createMod(v)
		err := apis.DownloadFile(config)
		if err != nil {
			return err
		}
	}
	return nil
}

func createMod(v string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; U; Linux x86_64; zh-CN; rv:1.9.2.10) "+
			"Gecko/20100922 Ubuntu/10.10 (maverick) Firefox/3.6.10")
		req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;")
		req.Header.Set("Origin", req.Host)
		req.Header.Set("Referer", v)
	}

}

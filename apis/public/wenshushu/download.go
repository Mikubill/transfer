package wenshushu

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"regexp"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

const (
	tokenConverter  = "https://www.wenshushu.cn/ap/task/token"
	downloadDetails = "https://www.wenshushu.cn/ap/task/mgrtask"
	downloadList    = "https://www.wenshushu.cn/ap/ufile/list"
	signDownload    = "https://www.wenshushu.cn/ap/dl/sign"
)

var (
	matcher  = regexp.MustCompile(`(https://)?ws+?[0-9]+?\\.cn/f/[0-9a-z]{11}|www\\.wenshushu\\.cn/t/[0-9a-zA-Z]{16}`)
	regex    = regexp.MustCompile("[0-9a-z]{11}")
	regexMgr = regexp.MustCompile("[0-9a-zA-Z]{16}")
)

func (b wssTransfer) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}

func (b wssTransfer) DoDownload(link string, config apis.DownConfig) error {
	err := b.download(link, config)
	if err != nil {
		return fmt.Errorf("download failed on %s, returns %s\n", link, err)
	}
	return nil
}

func (b wssTransfer) download(v string, config apis.DownConfig) error {
	ticket, err := b.getTicket()
	if err != nil {
		return err
	}

	var fileID string

	mgrID := regexMgr.FindString(v)
	if mgrID != "" {
		data, _ := json.Marshal(map[string]any{"token": mgrID})
		config, err := newRequest(tokenConverter, string(data), requestConfig{
			debug:    apis.DebugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(ticket),
		})
		if err != nil {
			return err
		}
		fileID = config.Data.Tid
	} else {
		fileID = regex.FindString(v)
	}
	//log.Println(fileID)

	if apis.DebugMode {
		log.Println("starting download...")
		log.Println("step1 -> api/getTicket")
	}
	fmt.Printf("Remote: %s\n", v)
	data, _ := json.Marshal(map[string]any{
		"tid":      fileID,
		"password": b.Config.passCode,
	})
	downConfig, err := newRequest(downloadDetails, string(data), requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	// log.Println(downConfig, err)
	if err != nil {
		return err
	}

	// todo: type 1/2, start(page?)
	data, _ = json.Marshal(map[string]any{
		"bid":     downConfig.Data.BoxID,
		"pid":     downConfig.Data.UFileID,
		"type":    1,
		"start":   0,
		"size":    50,
		"sort":    map[string]string{"name": "asc"},
		"options": map[string]string{"uploader": "true"},
	})
	downConfig, err = newRequest(downloadList, string(data), requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return err
	}

	for _, item := range downConfig.Data.FileList {
		err = b.downloadItem(item, ticket, config)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (b wssTransfer) downloadItem(item fileItem, token string, config apis.DownConfig) error {
	if apis.DebugMode {
		log.Println("step2 -> api/getConf")
		log.Printf("fileName: %s\n", item.FileName)
	}
	data, _ := json.Marshal(map[string]any{
		"consumeCode": 0,
		// "bid": item.Bid,
		"type":    "1",
		"ufileid": item.Fid,
	})

	resp, err := newRequest(signDownload, string(data), requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(token),
	})
	if err != nil {
		return fmt.Errorf("sign Request returns error: %s, onfile: %s", err, item.FileName)
	}

	if apis.DebugMode {
		log.Println("step3 -> startDownload")
	}
	filePath := config.Prefix

	if utils.IsExist(config.Prefix) {
		if utils.IsFile(config.Prefix) {
			filePath = config.Prefix
		} else {
			filePath = path.Join(config.Prefix, item.FileName)
		}
	}

	config.Prefix = filePath
	config.Link = resp.Data.URL
	config.Modifier = addHeaders

	err = apis.DownloadFile(config)
	if err != nil {
		return fmt.Errorf("failed DownloaderConfig with error: %s, onfile: %s", err, item.FileName)
	}
	return nil
}

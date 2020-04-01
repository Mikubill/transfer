package wenshushu

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"regexp"
	"time"
	"transfer/utils"
)

const (
	tokenConverter  = "https://www.wenshushu.cn/ap/task/token"
	downloadDetails = "https://www.wenshushu.cn/ap/task/mgrtask"
	downloadList    = "https://www.wenshushu.cn/ap/ufile/list"
	signDownload    = "https://www.wenshushu.cn/ap/dl/sign"
)

var (
	regex    = regexp.MustCompile("[0-9a-z]{11}")
	regexMgr = regexp.MustCompile("[0-9a-zA-Z]{16}")
)

func (b wssTransfer) Download(files []string) {
	for _, v := range files {
		err := b.download(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b wssTransfer) download(v string) error {
	ticket, err := b.getTicket()
	if err != nil {
		return err
	}

	var fileID string

	mgrID := regexMgr.FindString(v)
	if mgrID != "" {
		data, _ := json.Marshal(map[string]interface{}{"token": mgrID})
		config, err := newRequest(tokenConverter, string(data), requestConfig{
			debug:    b.Config.debugMode,
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
		if fileID == "" {
			return fmt.Errorf("unknown URL format")
		}
	}

	if b.Config.debugMode {
		log.Println("starting download...")
		log.Println("step1 -> api/getTicket")
	}
	fmt.Printf("Remote: %s\n", v)
	data, _ := json.Marshal(map[string]interface{}{
		"tid":      fileID,
		"password": b.Config.passCode,
	})
	config, err := newRequest(downloadDetails, string(data), requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return err
	}

	// todo: type 1/2, start(page?)
	data, _ = json.Marshal(map[string]interface{}{
		"bid":  config.Data.BoxID,
		"pid":  config.Data.UFileID,
		"type": 1,
	})
	config, err = newRequest(downloadList, string(data), requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return err
	}

	for _, item := range config.Data.FileList {
		err = b.downloadItem(item, ticket)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (b wssTransfer) downloadItem(item fileItem, token string) error {
	if b.Config.debugMode {
		log.Println("step2 -> api/getConf")
		log.Printf("fileName: %s\n", item.FileName)
	}
	data, _ := json.Marshal(map[string]interface{}{
		"bid": item.Bid,
		"fid": item.Fid,
	})

	resp, err := newRequest(signDownload, string(data), requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(token),
	})
	if err != nil {
		return fmt.Errorf("sign Request returns error: %s, onfile: %s", err, item.FileName)
	}

	if b.Config.debugMode {
		log.Println("step3 -> startDownload")
	}
	filePath := b.Config.prefix

	if utils.IsExist(b.Config.prefix) {
		if utils.IsFile(b.Config.prefix) {
			filePath = b.Config.prefix
		} else {
			filePath = path.Join(b.Config.prefix, item.FileName)
		}
	}

	//fmt.Printf("File save to: %s\n", filePath)
	err = utils.DownloadFile(filePath, resp.Data.URL, utils.DownloadConfig{
		Force:    b.Config.forceMode,
		Debug:    b.Config.debugMode,
		Parallel: b.Config.parallel,
		Modifier: addHeaders,
	})
	if err != nil {
		return fmt.Errorf("failed DownloadConfig with error: %s, onfile: %s", err, item.FileName)
	}
	return nil
}

package cowtransfer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"transfer/utils"
)

const (
	downloadDetails = "https://cowtransfer.com/transfer/transferdetail?url=%s&treceive=undefined&passcode=%s"
	downloadConfig  = "https://cowtransfer.com/transfer/download?guid=%s"
)

var regex = regexp.MustCompile("[0-9a-f]{14}")

func (b cowTransfer) Download(files []string) {
	for _, v := range files {
		err := b.initDownload(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b cowTransfer) initDownload(v string) error {
	fileID := regex.FindString(v)
	if fileID == "" {
		return fmt.Errorf("unknown URL format")
	}

	if b.Config.debugMode {
		log.Println("starting download...")
		log.Println("step1 -> api/getGuid")
	}
	fmt.Printf("Remote: %s\n", v)
	detailsURL := fmt.Sprintf(downloadDetails, fileID, b.Config.passCode)
	resp, err := http.Get(detailsURL)
	if err != nil {
		return fmt.Errorf("getDownloadDetails returns error: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("readDownloadDetails returns error: %s", err)
	}

	_ = resp.Body.Close()

	if b.Config.debugMode {
		log.Printf("returns: %v\n", string(body))
	}
	details := new(downloadDetailsResponse)
	if err := json.Unmarshal(body, details); err != nil {
		return fmt.Errorf("unmatshal DownloadDetails returns error: %s", err)
	}

	if details.GUID == "" {
		return fmt.Errorf("link invalid")
	}

	if details.Deleted {
		return fmt.Errorf("link deleted")
	}

	if !details.Uploaded {
		return fmt.Errorf("link not finish upload yet")
	}

	for _, item := range details.Details {
		err = b.downloadItem(item)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (b cowTransfer) downloadItem(item downloadDetailsBlock) error {
	if b.Config.debugMode {
		log.Println("step2 -> api/getConf")
		log.Printf("fileName: %s\n", item.FileName)
		log.Printf("fileSize: %s\n", item.Size)
		log.Printf("GUID: %s\n", item.GUID)
	}
	configURL := fmt.Sprintf(downloadConfig, item.GUID)
	req, err := http.NewRequest("POST", configURL, nil)
	if err != nil {
		return fmt.Errorf("createRequest returns error: %s, onfile: %s", err, item.FileName)
	}
	addHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("getDownloadConfig returns error: %s, onfile: %s", err, item.FileName)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("readDownloadConfig returns error: %s, onfile: %s", err, item.FileName)
	}

	_ = resp.Body.Close()
	if b.Config.debugMode {
		log.Printf("returns: %v\n", string(body))
	}
	config := new(downloadConfigResponse)
	if err := json.Unmarshal(body, config); err != nil {
		return fmt.Errorf("unmatshal DownloadConfig returns error: %s, onfile: %s", err, item.FileName)
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
	err = utils.DownloadFile(filePath, config.Link, utils.DownloadConfig{
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

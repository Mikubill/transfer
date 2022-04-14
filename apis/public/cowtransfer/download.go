package cowtransfer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

const (
	downloadDetails = "https://cowtransfer.com/api/transfer/v2/transferdetail?url=%s&treceive=undefined&passcode=%s"
	downloadFiles   = "https://cowtransfer.com/api/transfer/v2/files?page=0&guid=%s"
	downloadConfig  = "https://cowtransfer.com/api/transfer/download?guid=%s"
)

var (
	matcher = regexp.MustCompile(`(cowtransfer\.com|c-t\.work)/s/[0-9a-f]{14}`)
	reg     = regexp.MustCompile("[0-9a-f]{14}")
)

func (b cowTransfer) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}

func (b cowTransfer) DoDownload(link string, config apis.DownConfig) error {
	return b.initDownload(link, config)
}

func (b cowTransfer) initDownload(v string, config apis.DownConfig) error {
	fileID := reg.FindString(v)
	if apis.DebugMode {
		log.Println("starting download...")
		log.Println("step1 -> api/getGuid")
	}
	fmt.Printf("Remote: %s\n", v)

	body, err := fetchWithCookie(fmt.Sprintf(downloadDetails, fileID, config.Ticket), fileID)
	if err != nil {
		return fmt.Errorf("request DownloadDetails returns error: %s", err)
	}

	if apis.DebugMode {
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

	body, err = fetchWithCookie(fmt.Sprintf(downloadFiles, details.GUID), fileID)
	if err != nil {
		return fmt.Errorf("request FileDetails returns error: %s", err)
	}

	if apis.DebugMode {
		log.Printf("returns: %v\n", string(body))
	}

	files := new(downloadFilesResponse)
	if err := json.Unmarshal(body, files); err != nil {
		return fmt.Errorf("unmatshal DownloadDetails returns error: %s", err)
	}

	for _, item := range files.Details {
		err = downloadItem(item, config)
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchWithCookie(link, fileID string) ([]byte, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, fmt.Errorf("getDownloadDetails returns error: %s", err)
	}
	req.Header.Set("Referer", fmt.Sprintf("https://cowtransfer.com/s/%s", fileID))
	req.Header.Set("Cookie", fmt.Sprintf("cf-cs-k-20181214=%d;", time.Now().UnixNano()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getDownloadDetails returns error: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("readDownloadDetails returns error: %s", err)
	}

	_ = resp.Body.Close()
	return body, nil
}

func downloadItem(item downloadDetailsBlock, baseConf apis.DownConfig) error {
	if apis.DebugMode {
		log.Println("step2 -> api/getConf")
		log.Printf("fileName: %s\n", item.FileName)
		log.Printf("fileSize: %2.f\n", item.Size)
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
	if apis.DebugMode {
		log.Printf("returns: %v\n", string(body))
	}
	config := new(downloadConfigResponse)
	if err := json.Unmarshal(body, config); err != nil {
		return fmt.Errorf("unmatshal DownloaderConfig returns error: %s, onfile: %s", err, item.FileName)
	}

	if apis.DebugMode {
		log.Println("step3 -> startDownload")
	}
	filePath, err := filepath.Abs(baseConf.Prefix)
	if err != nil {
		return fmt.Errorf("invalid prefix: %s", baseConf.Prefix)
	}

	if utils.IsExist(baseConf.Prefix) {
		if utils.IsFile(baseConf.Prefix) {
			filePath = baseConf.Prefix
		} else {
			filePath = path.Join(baseConf.Prefix, item.FileName)
		}
	}

	baseConf.Prefix = filePath
	baseConf.Link = config.Link
	baseConf.Modifier = addHeaders

	err = apis.DownloadFile(baseConf)
	if err != nil {
		return fmt.Errorf("failed DownloaderConfig with error: %s, onfile: %s", err, item.FileName)
	}
	return nil
}

package cowtransfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"transfer/utils"
)

const (
	prepareSend        = "https://cowtransfer.com/transfer/preparesend"
	setPassword        = "https://cowtransfer.com/transfer/v2/bindpasscode"
	beforeUpload       = "https://cowtransfer.com/transfer/beforeupload"
	uploadInitEndpoint = "https://upload.qiniup.com/mkblk/%d"
	uploadEndpoint     = "https://upload.qiniup.com/bput/%s/%d"
	uploadFinish       = "https://cowtransfer.com/transfer/uploaded"
	uploadComplete     = "https://cowtransfer.com/transfer/complete"
	uploadMergeFile    = "https://upload.qiniup.com/mkfile/%s/key/%s/fname/%s"
	block              = 4194304
)

func (b cowTransfer) Upload(files []string) {
	if b.Config.singleMode {
		err := b.initUpload(files)
		if err != nil {
			fmt.Printf("upload failed on %s, returns %s\n", files, err)
		}
	} else {
		for _, v := range files {
			err := b.initUpload([]string{v})
			if err != nil {
				fmt.Printf("upload failed on %s, returns %s\n", v, err)
			}
		}
	}
}

func (b cowTransfer) initUpload(files []string) error {
	totalSize := int64(0)

	for _, v := range files {
		err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			totalSize += info.Size()
			return nil
		})
		if err != nil {
			fmt.Printf("filepath.walk returns error: %v, onfile: %s\n", err, v)
		}
	}

	config, err := b.getSendConfig(totalSize)
	if err != nil {
		fmt.Printf("getSendConfig(single mode) returns error: %v\n", err)
	}
	fmt.Printf("Destination: %s\n", config.UniqueURL)
	for _, v := range files {
		err = filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			err = b.upload(path, config)
			if err != nil {
				fmt.Printf("upload returns error: %v, onfile: %s\n", err, path)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("filepath.walk(upload) returns error: %v, onfile: %s\n", err, v)
		}
	}
	err = b.completeUpload(config)
	if err != nil {
		fmt.Printf("complete upload(single mode) returns error: %v\n", err)
	}
	return nil
}

func (b cowTransfer) upload(v string, baseConf *prepareSendResp) error {
	fmt.Printf("Local: %s\n", v)
	if b.Config.debugMode {
		log.Println("retrieving file info...")
	}
	info, err := utils.GetFileInfo(v)
	if err != nil {
		return fmt.Errorf("getFileInfo returns error: %v", err)
	}

	config, err := b.getUploadConfig(info, baseConf)
	if err != nil {
		return fmt.Errorf("getUploadConfig returns error: %v", err)
	}
	bar := pb.Full.Start64(info.Size())
	bar.Set(pb.Bytes, true)
	file, err := os.Open(v)
	if err != nil {
		return fmt.Errorf("openFile returns error: %v", err)
	}

	wg := new(sync.WaitGroup)
	ch := make(chan *uploadPart)
	hashMap := cmap.New()
	for i := 0; i < b.Config.parallel; i++ {
		go b.uploader(&ch, wg, bar, config.UploadToken, &hashMap)
	}
	part := int64(0)
	for {
		part++
		buf := make([]byte, block)
		nr, err := file.Read(buf)
		if nr <= 0 || err != nil {
			break
		}
		if nr > 0 {
			wg.Add(1)
			ch <- &uploadPart{
				content: buf[:nr],
				count:   part,
			}
		}
	}

	wg.Wait()
	close(ch)
	_ = file.Close()
	bar.Finish()
	// finish upload
	err = b.finishUpload(config, info, &hashMap, part)
	if err != nil {
		return fmt.Errorf("finishUpload returns error: %v", err)
	}
	return nil
}

func (b cowTransfer) uploader(ch *chan *uploadPart, wg *sync.WaitGroup, bar *pb.ProgressBar, token string, hashMap *cmap.ConcurrentMap) {
	for item := range *ch {
		postURL := fmt.Sprintf(uploadInitEndpoint, len(item.content))
		if b.Config.debugMode {
			log.Printf("part %d start uploading, size: %d", item.count, len(item.content))
			log.Printf("part %d posting %s", item.count, postURL)
		}

		// makeBlock
		body, err := newPostRequest(postURL, nil, requestConfig{
			debug:    b.Config.debugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(token),
		})
		if err != nil {
			if b.Config.debugMode {
				log.Printf("failed make mkblk on part %d, error: %s (retrying)",
					item.count, err)
			}
			*ch <- item
			continue
		}
		var rBody uploadResponse
		if err := json.Unmarshal(body, &rBody); err != nil {
			if b.Config.debugMode {
				log.Printf("failed make mkblk on part %d error: %v, returns: %s (retrying)",
					item.count, string(body), strings.ReplaceAll(err.Error(), "\n", ""))
			}
			*ch <- item
			continue
		}

		//blockPut
		failFlag := false
		blockCount := int(math.Ceil(float64(len(item.content)) / float64(b.Config.blockSize)))
		if b.Config.debugMode {
			log.Printf("init: part %d block %d ", item.count, blockCount)
		}
		ticket := rBody.Ticket
		for i := 0; i < blockCount; i++ {
			start := i * b.Config.blockSize
			end := (i + 1) * b.Config.blockSize
			var buf []byte
			if end > len(item.content) {
				buf = item.content[start:]
			} else {
				buf = item.content[start:end]
			}
			if b.Config.debugMode {
				log.Printf("part %d block %d [%d:%d] start upload...", item.count, i, start, end)
			}
			postURL = fmt.Sprintf(uploadEndpoint, ticket, start)
			ticket, err = b.blockPut(postURL, buf, token, 0)
			if err != nil {
				if b.Config.debugMode {
					log.Printf("part %d block %d failed. error: %s (retrying)", item.count, i, err)
				}
				failFlag = true
				break
			}
			bar.Add(len(buf))
		}
		if failFlag {
			*ch <- item
			continue
		}

		if b.Config.debugMode {
			log.Printf("part %d finished.", item.count)
		}
		hashMap.Set(strconv.FormatInt(item.count, 10), ticket)
		wg.Done()
	}

}

func (b cowTransfer) blockPut(postURL string, buf []byte, token string, retry int) (string, error) {
	data := new(bytes.Buffer)
	data.Write(buf)
	body, err := newPostRequest(postURL, data, requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(token),
	})
	if err != nil {
		if b.Config.debugMode {
			log.Printf("block upload failed (retrying)")
		}
		if retry > 3 {
			return "", err
		}
		return b.blockPut(postURL, buf, token, retry+1)
	}
	var rBody uploadResponse
	if err := json.Unmarshal(body, &rBody); err != nil {
		if b.Config.debugMode {
			log.Printf("block upload failed (retrying)")
		}
		if retry > 3 {
			return "", err
		}
		return b.blockPut(postURL, buf, token, retry+1)
	}
	if b.Config.hashCheck {
		if utils.HashBlock(buf) != rBody.Hash {
			if b.Config.debugMode {
				log.Printf("block hashcheck failed (retrying)")
			}
			if retry > 3 {
				return "", err
			}
			return b.blockPut(postURL, buf, token, retry+1)
		}
	}
	return rBody.Ticket, nil
}

func (b cowTransfer) finishUpload(config *prepareSendResp, info os.FileInfo, hashMap *cmap.ConcurrentMap, limit int64) error {
	if b.Config.debugMode {
		log.Println("finishing upload...")
		log.Println("step1 -> api/mergeFile")
	}
	filename := utils.URLSafeEncode(info.Name())
	fileLocate := utils.URLSafeEncode(fmt.Sprintf("anonymous/%s/%s", config.TransferGUID, info.Name()))
	mergeFileURL := fmt.Sprintf(uploadMergeFile, strconv.FormatInt(info.Size(), 10), fileLocate, filename)
	postBody := ""
	for i := int64(1); i <= limit; i++ {
		item, alimasu := hashMap.Get(strconv.FormatInt(i, 10))
		if alimasu {
			postBody += item.(string) + ","
		}
	}
	if strings.HasSuffix(postBody, ",") {
		postBody = postBody[:len(postBody)-1]
	}
	if b.Config.debugMode {
		log.Printf("merge payload: %s\n", postBody)
	}
	reader := bytes.NewReader([]byte(postBody))
	_, err := newPostRequest(mergeFileURL, reader, requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.UploadToken),
	})
	if err != nil {
		return err
	}

	if b.Config.debugMode {
		log.Println("step2 -> api/uploaded")
	}
	data := map[string]string{"transferGuid": config.TransferGUID, "fileId": ""}
	body, err := newMultipartRequest(uploadFinish, data, requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addHeaders,
	})
	if err != nil {
		return err
	}
	if string(body) != "true" {
		return fmt.Errorf("finish upload failed: status != true")
	}
	return nil
}

func (b cowTransfer) completeUpload(config *prepareSendResp) error {
	data := map[string]string{"transferGuid": config.TransferGUID, "fileId": ""}
	if b.Config.debugMode {
		log.Println("step3 -> api/completeUpload")
	}
	body, err := newMultipartRequest(uploadComplete, data, requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addHeaders,
	})
	if err != nil {
		return err
	}
	var rBody finishResponse
	if err := json.Unmarshal(body, &rBody); err != nil {
		return fmt.Errorf("read finish resp failed: %s", err)
	}
	if rBody.Status != true {
		return fmt.Errorf("finish upload failed: complete is not true")
	}
	fmt.Printf("Short Download Code: %s\n", rBody.TempDownloadCode)
	return nil
}

func (b cowTransfer) getSendConfig(totalSize int64) (*prepareSendResp, error) {
	data := map[string]string{
		"totalSize": strconv.FormatInt(totalSize, 10),
	}
	body, err := newMultipartRequest(prepareSend, data, requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addHeaders,
	})
	if err != nil {
		return nil, err
	}
	config := new(prepareSendResp)
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}
	if config.Error != false {
		return nil, fmt.Errorf(config.ErrorMessage)
	}
	if b.Config.passCode != "" {
		// set password
		data := map[string]string{
			"transferguid": config.TransferGUID,
			"passcode":     b.Config.passCode,
		}
		body, err = newMultipartRequest(setPassword, data, requestConfig{
			debug:    b.Config.debugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addHeaders,
		})
		if err != nil {
			return nil, err
		}
		if string(body) != "true" {
			return nil, fmt.Errorf("set password unsuccessful")
		}
	}
	return config, nil
}

func (b cowTransfer) getUploadConfig(info os.FileInfo, config *prepareSendResp) (*prepareSendResp, error) {

	if b.Config.debugMode {
		log.Println("retrieving upload config...")
		log.Println("step 2/2 -> beforeUpload")
	}

	data := map[string]string{
		"fileId":        "",
		"type":          "",
		"fileName":      info.Name(),
		"fileSize":      strconv.FormatInt(info.Size(), 10),
		"transferGuid":  config.TransferGUID,
		"storagePrefix": config.Prefix,
	}
	_, err := newMultipartRequest(beforeUpload, data, requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addHeaders,
	})
	if err != nil {
		return nil, err
	}
	return config, nil
}

func newPostRequest(link string, postBody io.Reader, config requestConfig) ([]byte, error) {
	if config.debug {
		log.Printf("retrying: %v", config.retry)
		log.Printf("endpoint: %s", link)
	}
	client := http.Client{}
	if config.timeout != 0 {
		client = http.Client{Timeout: config.timeout}
	}
	req, err := http.NewRequest("POST", link, postBody)
	if err != nil {
		if config.debug {
			log.Printf("build requests returns error: %v", err)
		}
		if config.retry > 3 {
			return nil, err
		}
		return newPostRequest(link, postBody, config)
	}
	config.modifier(req)
	resp, err := client.Do(req)
	if err != nil {
		if config.debug {
			log.Printf("do requests returns error: %v", err)
		}
		if config.retry > 3 {
			return nil, err
		}
		config.retry++
		return newPostRequest(link, postBody, config)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", err)
		}
		if config.retry > 3 {
			return nil, err
		}
		config.retry++
		return newPostRequest(link, postBody, config)
	}
	_ = resp.Body.Close()
	if config.debug {
		if len(body) < 1024 {
			log.Printf("returns: %v", string(body))
		}
	}
	return body, nil
}

func newMultipartRequest(url string, params map[string]string, config requestConfig) ([]byte, error) {
	if config.debug {
		log.Printf("retrying: %v", config.retry)
		log.Printf("endpoint: %s", url)
	}
	client := http.Client{}
	if config.timeout != 0 {
		client = http.Client{Timeout: config.timeout}
	}
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	_ = writer.Close()
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		if config.debug {
			log.Printf("build requests returns error: %v", err)
		}
		if config.retry > 3 {
			return nil, err
		}
		config.retry++
		return newMultipartRequest(url, params, config)
	}
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	config.modifier(req)
	if config.debug {
		log.Printf("header: %v", req.Header)
	}
	resp, err := client.Do(req)
	if err != nil {
		if config.debug {
			log.Printf("do requests returns error: %v", err)
		}
		if config.retry > 3 {
			return nil, err
		}
		config.retry++
		return newMultipartRequest(url, params, config)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", err)
		}
		if config.retry > 3 {
			return nil, err
		}
		config.retry++
		return newMultipartRequest(url, params, config)
	}
	_ = resp.Body.Close()
	if config.debug {
		log.Printf("returns: %v", string(body))
	}

	return body, nil
}

func addToken(upToken string) func(req *http.Request) {
	return func(req *http.Request) {
		addHeaders(req)
		req.Header.Set("Authorization", "UpToken "+upToken)
	}
}

func addHeaders(req *http.Request) {
	req.Header.Set("Referer", "https://cowtransfer.com/")
	req.Header.Set("User-Agent", "Chrome/80.0.3987.149 CowTransfer-Uploader")
	req.Header.Set("Origin", "https://cowtransfer.com/")
}

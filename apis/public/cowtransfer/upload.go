package cowtransfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"transfer/apis"
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

func (b cowTransfer) InitUpload(_ []string, sizes []int64) error {
	if b.Config.singleMode {
		totalSize := int64(0)
		for _, v := range sizes {
			totalSize += v
		}
		err := b.initUpload(totalSize)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *cowTransfer) initUpload(totalSize int64) error {
	config, err := b.getSendConfig(totalSize)
	if err != nil {
		fmt.Printf("getSendConfig(single mode) returns error: %v\n", err)
	}
	fmt.Printf("Destination: %s\n", config.UniqueURL)
	b.sendConf = *config
	return nil
}

func (b *cowTransfer) PreUpload(_ string, size int64) error {
	if !b.Config.singleMode {
		err := b.initUpload(size)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *cowTransfer) StartProgress(reader io.Reader, size int64) io.Reader {
	bar := pb.Full.Start64(size)
	bar.Set(pb.Bytes, true)
	b.Bar = bar
	return reader
}

func (b cowTransfer) DoUpload(name string, size int64, file io.Reader) error {

	config, err := b.getUploadConfig(name, size, b.sendConf)
	if err != nil {
		return fmt.Errorf("getUploadConfig returns error: %v", err)
	}

	wg := new(sync.WaitGroup)
	ch := make(chan *uploadPart)
	hashMap := cmap.New()
	for i := 0; i < b.Config.Parallel; i++ {
		go b.uploader(&ch, uploadConfig{
			wg:      wg,
			token:   config.UploadToken,
			hashMap: &hashMap,
		})
	}

	part := int64(0)
	for {
		part++
		buf := make([]byte, block)
		nr, err := io.ReadFull(file, buf)
		if nr <= 0 {
			break
		}
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Println(err)
			break
		}
		if nr > 0 {
			wg.Add(1)
			ch <- &uploadPart{
				content: buf[:nr],
				count:   part,
				bar:     b.Bar,
			}
		}
	}

	wg.Wait()
	close(ch)
	// finish upload
	err = b.finishUpload(config, name, size, &hashMap, part)
	if err != nil {
		return fmt.Errorf("finishUpload returns error: %v", err)
	}
	return nil
}

func (b cowTransfer) uploader(ch *chan *uploadPart, conf uploadConfig) {
	for item := range *ch {
		postURL := fmt.Sprintf(uploadInitEndpoint, len(item.content))
		if apis.DebugMode {
			log.Printf("part %d start uploading, size: %d", item.count, len(item.content))
			log.Printf("part %d posting %s", item.count, postURL)
		}

		// makeBlock
		body, err := newPostRequest(postURL, nil, requestConfig{
			debug:    apis.DebugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(conf.token),
		})
		if err != nil {
			if apis.DebugMode {
				log.Printf("failed make mkblk on part %d, error: %s (retrying)",
					item.count, err)
			}
			*ch <- item
			continue
		}
		var rBody uploadResponse
		if err := json.Unmarshal(body, &rBody); err != nil {
			if apis.DebugMode {
				log.Printf("failed make mkblk on part %d error: %v, returns: %s (retrying)",
					item.count, string(body), strings.ReplaceAll(err.Error(), "\n", ""))
			}
			*ch <- item
			continue
		}

		//blockPut
		failFlag := false
		blockCount := int(math.Ceil(float64(len(item.content)) / float64(b.Config.blockSize)))
		if apis.DebugMode {
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
			if apis.DebugMode {
				log.Printf("part %d block %d [%d:%d] start upload...", item.count, i, start, end)
			}
			postURL = fmt.Sprintf(uploadEndpoint, ticket, start)
			ticket, err = b.blockPut(postURL, buf, conf.token, 0)
			if err != nil {
				if apis.DebugMode {
					log.Printf("part %d block %d failed. error: %s (retrying)", item.count, i, err)
				}
				failFlag = true
				break
			}
			if item.bar != nil {
				item.bar.Add(len(buf))
			}
		}
		if failFlag {
			*ch <- item
			continue
		}

		if apis.DebugMode {
			log.Printf("part %d finished.", item.count)
		}
		conf.hashMap.Set(strconv.FormatInt(item.count, 10), ticket)
		conf.wg.Done()
	}

}

func (b cowTransfer) blockPut(postURL string, buf []byte, token string, retry int) (string, error) {
	data := new(bytes.Buffer)
	data.Write(buf)
	body, err := newPostRequest(postURL, data, requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(token),
	})
	if err != nil {
		if apis.DebugMode {
			log.Printf("block upload failed (retrying)")
		}
		if retry > 3 {
			return "", err
		}
		return b.blockPut(postURL, buf, token, retry+1)
	}
	var rBody uploadResponse
	if err := json.Unmarshal(body, &rBody); err != nil {
		if apis.DebugMode {
			log.Printf("block upload failed (retrying)")
		}
		if retry > 3 {
			return "", err
		}
		return b.blockPut(postURL, buf, token, retry+1)
	}
	if b.Config.hashCheck {
		if hashBlock(buf) != rBody.Hash {
			if apis.DebugMode {
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

func hashBlock(buf []byte) int {
	return int(crc32.ChecksumIEEE(buf))
}

func (b cowTransfer) finishUpload(config *prepareSendResp, name string, size int64, hashMap *cmap.ConcurrentMap, limit int64) error {
	if apis.DebugMode {
		log.Println("finishing upload...")
		log.Println("step1 -> api/mergeFile")
	}
	filename := utils.URLSafeEncode(name)
	fileLocate := utils.URLSafeEncode(fmt.Sprintf("%s/%s/%s", config.Prefix, config.TransferGUID, name))
	mergeFileURL := fmt.Sprintf(uploadMergeFile, strconv.FormatInt(size, 10), fileLocate, filename)
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
	if apis.DebugMode {
		log.Printf("merge payload: %s\n", postBody)
	}
	reader := bytes.NewReader([]byte(postBody))
	_, err := newPostRequest(mergeFileURL, reader, requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.UploadToken),
	})
	if err != nil {
		return err
	}

	if apis.DebugMode {
		log.Println("step2 -> api/uploaded")
	}
	data := map[string]string{"transferGuid": config.TransferGUID, "fileId": ""}
	body, err := b.newMultipartRequest(uploadFinish, data, requestConfig{
		debug:    apis.DebugMode,
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

func (b cowTransfer) FinishUpload([]string) error {
	data := map[string]string{"transferGuid": b.sendConf.TransferGUID, "fileId": ""}
	if apis.DebugMode {
		log.Println("step3 -> api/completeUpload")
	}
	body, err := b.newMultipartRequest(uploadComplete, data, requestConfig{
		debug:    apis.DebugMode,
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
	body, err := b.newMultipartRequest(prepareSend, data, requestConfig{
		debug:    apis.DebugMode,
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
		body, err = b.newMultipartRequest(setPassword, data, requestConfig{
			debug:    apis.DebugMode,
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

func (b cowTransfer) getUploadConfig(name string, size int64, config prepareSendResp) (*prepareSendResp, error) {

	if apis.DebugMode {
		log.Println("retrieving upload config...")
		log.Println("step 2/2 -> beforeUpload")
	}

	data := map[string]string{
		"fileId":        "",
		"type":          "",
		"fileName":      name,
		"fileSize":      strconv.FormatInt(size, 10),
		"transferGuid":  config.TransferGUID,
		"storagePrefix": config.Prefix,
	}
	_, err := b.newMultipartRequest(beforeUpload, data, requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addHeaders,
	})
	if err != nil {
		return nil, err
	}
	return &config, nil
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

func (b cowTransfer) newMultipartRequest(url string, params map[string]string, config requestConfig) ([]byte, error) {
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
		return b.newMultipartRequest(url, params, config)
	}
	req.Header.Set("cookie", b.Config.token)
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
		return b.newMultipartRequest(url, params, config)
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
		return b.newMultipartRequest(url, params, config)
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

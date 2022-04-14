package cowtransfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	prepareSend    = "https://cowtransfer.com/api/transfer/v2/preparesend"
	setPassword    = "https://cowtransfer.com/api/transfer/v2/bindpasscode"
	beforeUpload   = "https://cowtransfer.com/api/transfer/v2/beforeupload"
	uploadFinish   = "https://cowtransfer.com/api/transfer/v2/uploaded"
	uploadComplete = "https://cowtransfer.com/api/transfer/v2/complete"
	initUpload     = "https://upload.qiniup.com/buckets/cftransfer/objects/%s/uploads"
	doUpload       = "https://upload.qiniup.com/buckets/cftransfer/objects/%s/uploads/%s/%d"
	finUpload      = "https://upload.qiniup.com/buckets/cftransfer/objects/%s/uploads/%s"
)

func (b *cowTransfer) InitUpload(_ []string, sizes []int64) error {
	if b.Config.singleMode {
		totalSize := int64(0)
		for _, v := range sizes {
			totalSize += v
		}
		return b.initUpload(totalSize)
	}
	return nil
}

func (b *cowTransfer) initUpload(totalSize int64) error {
	config, err := b.getSendConfig(totalSize)
	if err != nil {
		return err
	}
	b.sendConf = *config
	return nil
}

func (b *cowTransfer) PreUpload(_ string, size int64) error {
	if !b.Config.singleMode {
		return b.initUpload(size)
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
		return fmt.Errorf("getUploadConfig error: %v", err)
	}

	wg := new(sync.WaitGroup)
	ch := make(chan *uploadPart)
	hashMap := cmap.New()
	for i := 0; i < b.Config.Parallel; i++ {
		go b.uploader(&ch, uploadConfig{
			wg:      wg,
			config:  config,
			hashMap: &hashMap,
		})
	}

	partCount := size / int64(b.Config.blockSize)
	if partCount >= 10000 {
		b.Config.blockSize = size / int64(9990)
	}

	if b.Config.blockSize < 1200000 {
		b.Config.blockSize = 1200000
	}

	part := int64(0)
	for {
		part++
		buf := make([]byte, b.Config.blockSize)
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
	retry := 0
	for item := range *ch {
	Start:
		retry++
		if retry > 10 {
			fmt.Printf("Upload part %d failed after 10 retries\n", item.count)
			// temp fix issue #52, should be handled by uploader
			os.Exit(1)
		}

		postURL := fmt.Sprintf(doUpload, conf.config.EncodeID, conf.config.ID, item.count)
		if apis.DebugMode {
			log.Printf("part %d start uploading, size: %d", item.count, len(item.content))
			log.Printf("part %d posting %s", item.count, postURL)
		}

		//blockPut
		ticket, err := b.blockPut(postURL, item.content, conf.config.Token)
		if err != nil {
			if apis.DebugMode {
				log.Printf("part %d failed. error: %s", item.count, err)
			}
			goto Start
		}

		if apis.DebugMode {
			log.Printf("part %d finished.", item.count)
		}
		conf.hashMap.Set(strconv.FormatInt(item.count, 10), ticket)
		conf.wg.Done()
		if item.bar != nil {
			item.bar.Add(len(item.content))
		}
	}

}

func (b cowTransfer) finishUpload(config *initResp, name string, size int64, hashMap *cmap.ConcurrentMap, limit int64) error {
	if apis.DebugMode {
		log.Println("finishing upload...")
		log.Println("step1 -> api/mergeFile")
	}
	mergeFileURL := fmt.Sprintf(finUpload, config.EncodeID, config.ID)
	var postData clds
	for i := int64(1); i <= limit; i++ {
		item, alimasu := hashMap.Get(strconv.FormatInt(i, 10))
		if alimasu {
			postData.Parts = append(postData.Parts, slek{
				ETag: item.(string),
				Part: i,
			})
		}
	}
	postData.FName = name
	postBody, err := json.Marshal(postData)
	if err != nil {
		return err
	}
	if apis.DebugMode {
		log.Printf("merge payload: %s\n", postBody)
	}
	reader := bytes.NewReader([]byte(postBody))
	resp, err := newRequest(mergeFileURL, reader, requestConfig{
		debug:  apis.DebugMode,
		action: "POST",
		//retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.Token),
	})
	if err != nil {
		return err
	}

	// read returns
	var mergeResp *uploadResult
	if err = json.Unmarshal(resp, &mergeResp); err != nil {
		return err
	}

	if apis.DebugMode {
		log.Println("step2 -> api/uploaded")
	}
	data := map[string]string{
		"transferGuid": config.TransferGUID,
		"fileGuid":     config.FileGUID,
		"hash":         mergeResp.Hash,
	}
	body, err := b.newMultipartRequest(uploadFinish, data, requestConfig{
		debug: apis.DebugMode,
		//retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: b.addTk,
	})
	if err != nil {
		return err
	}
	if string(body) != "true" || bytes.Contains(body, []byte("error")) {
		return fmt.Errorf("finish upload failed: %s", body)
	}
	return nil
}

func (b cowTransfer) PostUpload(string, int64) (string, error) {
	if !b.Config.singleMode {
		return b.CompleteUpload()
	}
	return "", nil
}

func (b cowTransfer) FinishUpload([]string) (string, error) {
	if b.Config.singleMode {
		return b.CompleteUpload()
	}
	return "", nil
}

func (b cowTransfer) CompleteUpload() (string, error) {
	code, err := b.completeUpload()
	if err != nil {
		return "", err
	}
	fmt.Printf("Download Link: %s\nDownload Code: %s\n", b.sendConf.UniqueURL, code)
	return b.sendConf.UniqueURL, nil
}

func (b cowTransfer) completeUpload() (string, error) {
	data := map[string]string{"transferGuid": b.sendConf.TransferGUID, "fileId": ""}
	if apis.DebugMode {
		log.Println("step3 -> api/completeUpload")
	}
	body, err := b.newMultipartRequest(uploadComplete, data, requestConfig{
		debug: apis.DebugMode,
		//retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: b.addTk,
	})
	if err != nil {
		return "", err
	}
	var rBody finishResponse
	if err := json.Unmarshal(body, &rBody); err != nil {
		return "", fmt.Errorf("read finish resp failed: %s", err)
	}
	if !rBody.Status {
		return "", fmt.Errorf("finish upload failed: complete is not true")
	}
	//fmt.Printf("Short Download Code: %s\n\n", rBody.TempDownloadCode)
	return rBody.TempDownloadCode, nil
}

func (b cowTransfer) getSendConfig(totalSize int64) (*prepareSendResp, error) {
	data := map[string]string{
		"validDays":      strconv.Itoa(b.Config.validDays),
		"totalSize":      strconv.FormatInt(totalSize, 10),
		"enableDownload": "true",
		"enablePreview":  "true",
	}
	body, err := b.newMultipartRequest(prepareSend, data, requestConfig{
		debug: apis.DebugMode,
		//retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: b.addTk,
	})
	if err != nil {
		return nil, err
	}
	config := new(prepareSendResp)
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}
	if config.Error {
		return nil, fmt.Errorf(config.ErrorMessage)
	}
	if b.Config.passCode != "" {
		// set password
		data := map[string]string{
			"transferguid": config.TransferGUID,
			"passcode":     b.Config.passCode,
		}
		body, err = b.newMultipartRequest(setPassword, data, requestConfig{
			debug: apis.DebugMode,
			//retry:    0,
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

func (b cowTransfer) getUploadConfig(name string, size int64, config prepareSendResp) (*initResp, error) {

	if apis.DebugMode {
		log.Println("retrieving upload config...")
		log.Println("step 2/2 -> beforeUpload")
	}

	data := map[string]string{
		"fileId":        "",
		"type":          "",
		"fileName":      name,
		"originalName":  name,
		"fileSize":      strconv.FormatInt(size, 10),
		"transferGuid":  config.TransferGUID,
		"storagePrefix": config.Prefix,
	}
	resp, err := b.newMultipartRequest(beforeUpload, data, requestConfig{
		debug: apis.DebugMode,
		//retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: b.addTk,
	})
	if err != nil {
		return nil, err
	}
	var beforeResp *beforeSendResp
	if err = json.Unmarshal(resp, &beforeResp); err != nil {
		return nil, err
	}
	config.FileGUID = beforeResp.FileGuid

	data = map[string]string{
		"transferGuid":  config.TransferGUID,
		"storagePrefix": config.Prefix,
	}
	p, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	w := utils.URLSafeEncode(fmt.Sprintf("%s/%s/%s", config.Prefix, config.TransferGUID, name))
	inits := fmt.Sprintf(initUpload, w)
	resp, err = newRequest(inits, bytes.NewReader(p), requestConfig{
		debug:  apis.DebugMode,
		action: "POST",
		//retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.UploadToken),
	})
	if err != nil {
		return nil, err
	}
	var initResp *initResp
	if err = json.Unmarshal(resp, &initResp); err != nil {
		return nil, err
	}
	initResp.Token = config.UploadToken
	initResp.EncodeID = w
	initResp.TransferGUID = config.TransferGUID
	initResp.FileGUID = config.FileGUID

	// return config, nil
	return initResp, nil
}

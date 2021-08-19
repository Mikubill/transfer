package cowtransfer

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"
	"time"
	"transfer/apis"
	"transfer/utils"

	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	prepareSend    = "https://cowtransfer.com/api/transfer/preparesend"
	setPassword    = "https://cowtransfer.com/api/transfer/v2/bindpasscode"
	beforeUpload   = "https://cowtransfer.com/api/transfer/beforeupload"
	uploadFinish   = "https://cowtransfer.com/api/transfer/uploaded"
	uploadComplete = "https://cowtransfer.com/api/transfer/complete"
	initUpload     = "https://upload-fog-cn-east-1.qiniup.com/buckets/cowtransfer-yz/objects/%s/uploads"
	doUpload       = "https://upload-fog-cn-east-1.qiniup.com/buckets/cowtransfer-yz/objects/%s/uploads/%s/%d"
	finUpload      = "https://upload-fog-cn-east-1.qiniup.com/buckets/cowtransfer-yz/objects/%s/uploads/%s"
)

func (b *cowTransfer) InitUpload(_ []string, sizes []int64) error {
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
		return err
	}
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
	for item := range *ch {
	Start:
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

func (b cowTransfer) blockPut(postURL string, buf []byte, token string) (string, error) {
	data := new(bytes.Buffer)
	data.Write(buf)
	body, err := newRequest(postURL, data, requestConfig{
		debug:  apis.DebugMode,
		action: "PUT",

		//retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(token),
	})
	if err != nil {
		if apis.DebugMode {
			log.Printf("block upload failed (retrying)")
		}
		//if retry > 3 {
		return "", err
		//}
		//return b.blockPut(postURL, buf, token, retry+1)
	}
	var rBody uploadResponse
	if err := json.Unmarshal(body, &rBody); err != nil {
		if apis.DebugMode {
			log.Printf("resp unmarshal failed (retrying)")
		}
		//if retry > 3 {
		return "", err
		//}
		//return b.blockPut(postURL, buf, token, retry+1)
	}
	if b.Config.hashCheck {
		if hashBlock(buf) != rBody.MD5 {
			if apis.DebugMode {
				log.Printf("block hashcheck failed (retrying)")
			}
			//if retry > 3 {
			return "", err
			//}
			//return b.blockPut(postURL, buf, token, retry+1)
		}
	}
	if rBody.Error != "" {
		return "", fmt.Errorf(rBody.Error)
	}
	return rBody.Etag, nil
}

func hashBlock(buf []byte) string {
	return fmt.Sprintf("%x", md5.Sum(buf))
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
		modifier: addHeaders,
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
		modifier: addHeaders,
	})
	if err != nil {
		return "", err
	}
	var rBody finishResponse
	if err := json.Unmarshal(body, &rBody); err != nil {
		return "", fmt.Errorf("read finish resp failed: %s", err)
	}
	if rBody.Status != true {
		return "", fmt.Errorf("finish upload failed: complete is not true")
	}
	//fmt.Printf("Short Download Code: %s\n\n", rBody.TempDownloadCode)
	return rBody.TempDownloadCode, nil
}

func (b cowTransfer) getSendConfig(totalSize int64) (*prepareSendResp, error) {
	data := map[string]string{
		"totalSize": strconv.FormatInt(totalSize, 10),
	}
	body, err := b.newMultipartRequest(prepareSend, data, requestConfig{
		debug: apis.DebugMode,
		//retry:    0,
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
		modifier: addHeaders,
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

func newRequest(link string, postBody io.Reader, config requestConfig) ([]byte, error) {
	if config.debug {
		//if config.retry != 0 {
		//	log.Printf("retrying: %v", config.retry)
		//}
		log.Printf("endpoint: %s", link)
	}
	client := http.Client{}
	if config.timeout != 0 {
		client = http.Client{Timeout: config.timeout}
	}
	req, err := http.NewRequest(config.action, link, postBody)
	if err != nil {
		if config.debug {
			log.Printf("build requests returns error: %v", err)
		}
		//if config.retry > 3 {
		return nil, err
		//}
		//return newPostRequest(link, postBody, config)
	}
	config.modifier(req)
	resp, err := client.Do(req)
	if err != nil {
		if config.debug {
			log.Printf("do requests returns error: %v", err)
		}
		//if config.retry > 20 {
		return nil, err
		//}
		//config.retry++
		//return newPostRequest(link, postBody, config)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", err)
		}
		//if config.retry > 20 {
		return nil, err
		//}
		//config.retry++
		//return newPostRequest(link, postBody, config)
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
		//log.Printf("retrying: %v", config.retry)
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
		//if config.retry > 3 {
		return nil, err
		//}
		//config.retry++
		//time.Sleep(1)
		//return b.newMultipartRequest(url, params, config)
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
		//if config.retry > 3 {
		return nil, err
		//}
		//config.retry++
		//time.Sleep(1)
		//return b.newMultipartRequest(url, params, config)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", err)
		}
		//if config.retry > 3 {
		return nil, err
		//}
		//config.retry++
		//time.Sleep(1)
		//return b.newMultipartRequest(url, params, config)
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
	req.Header.Set("User-Agent", "Chrome/80.0.3987.149 Transfer")
	req.Header.Set("Origin", "https://cowtransfer.com/")
	req.Header.Set("Cookie", fmt.Sprintf("%scf-cs-k-20181214=%d;", req.Header.Get("Cookie"), time.Now().UnixNano()))
}

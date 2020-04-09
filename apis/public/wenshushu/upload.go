package wenshushu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"transfer/apis"
)

const (
	anonymous = "https://www.wenshushu.cn/ap/login/anonymous"
	addSend   = "https://www.wenshushu.cn/ap/task/addsend"
	getUpID   = "https://www.wenshushu.cn/ap/uploadv2/getupid"
	getUpURL  = "https://www.wenshushu.cn/ap/uploadv2/psurl"
	complete  = "https://www.wenshushu.cn/ap/uploadv2/complete"
	process   = "https://www.wenshushu.cn/ap/ufile/getprocess"
	finish    = "https://www.wenshushu.cn/ap/task/copysend"
)

func (b *wssTransfer) InitUpload(_ []string, sizes []int64) error {
	if b.Config.singleMode {
		totalSize := int64(0)
		for _, v := range sizes {
			totalSize += v
		}
		b.initUpload(totalSize, len(sizes))
	}
	return nil
}

func (b *wssTransfer) initUpload(totalSize int64, totalCount int) {

	config, err := b.getSendConfig(totalSize, totalCount)
	if err != nil {
		fmt.Printf("getSendConfig(single mode) returns error: %v\n", err)
	}
	b.baseConf = *config
}

func (b *wssTransfer) PreUpload(_ string, size int64) error {
	if !b.Config.singleMode {
		b.initUpload(size, 1)
	}
	return nil
}

func (b wssTransfer) DoUpload(name string, size int64, file io.Reader) error {

	if size/int64(b.Config.blockSize) > 10000 {
		b.Config.blockSize = int(size / 10000)
		fmt.Printf("blocksize too small, set to %d\n", b.Config.blockSize)
	}

	wg := new(sync.WaitGroup)
	ch := make(chan *uploadPart)
	for i := 0; i < b.Config.Parallel; i++ {
		go b.uploader(&ch, b.baseConf)
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
				name:    name,
				wg:      wg,
			}
		}
	}

	wg.Wait()
	close(ch)
	// finish upload
	err := b.finishUpload(b.baseConf, name)
	if err != nil {
		return fmt.Errorf("finishUpload returns error: %v", err)
	}

	return nil
}

func (b wssTransfer) PostUpload(string, int64) (string, error) {
	if !b.Config.singleMode {
		return b.completeUpload(b.baseConf)
	}
	return "", nil
}

func (b wssTransfer) FinishUpload([]string) (string, error) {
	if b.Config.singleMode {
		return b.completeUpload(b.baseConf)
	}
	return "", nil
}

func (b wssTransfer) uploader(ch *chan *uploadPart, config sendConfigBlock) {
	for item := range *ch {
		d, _ := json.Marshal(map[string]interface{}{
			"ispart": true,
			"fname":  item.name,
			"partnu": item.count,
			"fsize":  b.Config.blockSize,
			"upId":   config.UploadID,
		})
		uploadTicket, err := newRequest(getUpURL, string(d), requestConfig{
			debug:    apis.DebugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(config.Token),
		})
		if err != nil {
			if apis.DebugMode {
				log.Printf("get upload url request returns error: %v", err)
			}
			*ch <- item
			continue
		}

		client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
		data := new(bytes.Buffer)
		data.Write(item.content)
		if apis.DebugMode {
			log.Printf("part %d start uploading", item.count)
			log.Printf("part %d posting %s", item.count, uploadTicket.Data.URL)
		}
		req, err := http.NewRequest("PUT", uploadTicket.Data.URL, data)
		if err != nil {
			if apis.DebugMode {
				log.Printf("build request returns error: %v", err)
			}
			*ch <- item
			continue
		}
		req.Header.Set("content-type", "application/octet-stream")
		resp, err := client.Do(req)
		if err != nil {
			if apis.DebugMode {
				log.Printf("failed uploading part %d error: %v (retrying)", item.count, err)
			}
			*ch <- item
			continue
		}
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			if apis.DebugMode {
				log.Printf("failed uploading part %d error: %v (retrying)", item.count, err)
			}
			*ch <- item
			continue
		}

		_ = resp.Body.Close()

		if apis.DebugMode {
			log.Printf("part %d finished.", item.count)
		}
		item.wg.Done()
	}

}

func (b wssTransfer) finishUpload(config sendConfigBlock, name string) error {
	if apis.DebugMode {
		log.Println("finish upload...")
		log.Println("step1 -> complete")
	}
	d, _ := json.Marshal(map[string]interface{}{
		"ispart": true,
		"fname":  name,
		"location": map[string]string{
			"boxid": config.Bid,
			"preid": config.UFileID,
		},
		"upId": config.UploadID,
	})

	body, err := newRequest(complete, string(d), requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.Token),
	})
	if err != nil {
		return err
	}
	if body.Message != "success" {
		return fmt.Errorf("upload failed returns: %s", body.Message)
	}
	return nil
}

func (b wssTransfer) completeUpload(config sendConfigBlock) (string, error) {
	if apis.DebugMode {
		log.Println("complete upload...")
		log.Println("step1 -> process")
	}
	d, _ := json.Marshal(map[string]string{"processId": config.UploadID})
	for {
		body, err := newRequest(process, string(d), requestConfig{
			debug:    apis.DebugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(config.Token),
		})
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if body.Data.R == "success" {
			break
		}
		time.Sleep(time.Second)
	}

	if apis.DebugMode {
		log.Println("step2 -> finish(copySend)")
	}
	d, _ = json.Marshal(map[string]string{
		"bid":     config.Bid,
		"ufileid": config.UFileID,
		"tid":     config.Tid,
	})
	body, err := newRequest(finish, string(d), requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.Token),
	})
	if err != nil {
		return "", err
	}
	if body.Message != "success" {
		return "", fmt.Errorf("status != success")
	}
	fmt.Printf("Manage Link: %s\nDownload Link: %s\n", body.Data.ManageURL, body.Data.PublicURL)
	return body.Data.PublicURL, nil
}

func (b wssTransfer) getTicket() (string, error) {
	if b.Config.token != "" {
		return b.Config.token, nil
	}
	if apis.DebugMode {
		log.Println("getToken...")
	}
	config, err := newRequest(anonymous, "{\"dev_info\":\"{}\"}", requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(""),
	})
	if err != nil {
		return "", err
	}
	return config.Data.Token, nil
}

func (b wssTransfer) getSendConfig(totalSize int64, totalCount int) (*sendConfigBlock, error) {
	ticket, err := b.getTicket()
	if err != nil {
		return nil, err
	}

	if apis.DebugMode {
		log.Println("step 1/2 addSend")
	}
	data, _ := json.Marshal(map[string]interface{}{
		"sender":      "",
		"remark":      "",
		"isextension": false,
		"pwd":         "",
		"expire":      2,
		"recvs":       []string{"social", "public"},
		"file_size":   strconv.FormatInt(totalSize, 10),
		"file_count":  totalCount,
	})
	config, err := newRequest(addSend, string(data), requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return nil, err
	}

	if apis.DebugMode {
		log.Println("step 2/2 getUpID")
	}
	data, _ = json.Marshal(map[string]interface{}{
		"boxid":      config.Data.Bid,
		"preid":      config.Data.UFileID,
		"linkid":     config.Data.Tid,
		"utype":      "sendcopy",
		"originUpid": "",
		"length":     totalSize,
		"count":      totalCount,
	})
	upData, err := newRequest(getUpID, string(data), requestConfig{
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return nil, err
	}
	config.Data.UploadID = upData.Data.UploadID
	config.Data.Token = ticket
	if apis.DebugMode {
		log.Printf("%+v", config.Data)
	}
	return &config.Data, nil
}

func newRequest(link string, postBody string, config requestConfig) (*sendConfigResp, error) {
	if config.debug {
		log.Printf("endpoint: %s", link)
		log.Printf("postBody: %s", postBody)

	}

	client := http.Client{Timeout: config.timeout}
	req, err := http.NewRequest("POST", link, strings.NewReader(postBody))
	if err != nil {
		if config.debug {
			log.Printf("build request returns error: %v", err)
		}
		if config.retry > 3 {
			return nil, fmt.Errorf("request %s returns error: %s", link, err)
		}
		config.retry++
		return newRequest(link, postBody, config)
	}
	config.modifier(req)
	resp, err := client.Do(req)
	if err != nil {
		if config.debug {
			log.Printf("do request returns error: %v", err)
		}
		if config.retry > 3 {
			return nil, fmt.Errorf("post %s returns error: %s", link, err)
		}
		config.retry++
		return newRequest(link, postBody, config)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", err)
		}
		if config.retry > 3 {
			return nil, fmt.Errorf("post %s returns error: %s", link, err)
		}
		config.retry++
		return newRequest(link, postBody, config)
	}
	_ = resp.Body.Close()
	if config.debug {
		log.Printf("returns: %v", string(body))
	}

	respDat := new(sendConfigResp)
	err = json.Unmarshal(body, respDat)
	if err != nil || respDat.Message != "success" {
		if config.retry > 3 {
			return nil, fmt.Errorf("post %s returns error: %s", link, err)
		}
		config.retry++
		return newRequest(link, postBody, config)
	}
	if config.debug {
		log.Printf("%+v", respDat)
	}
	return respDat, nil
}

func addToken(token string) func(req *http.Request) {
	return func(req *http.Request) {
		addHeaders(req)
		req.Header.Set("x-token", token)
	}
}

func addHeaders(req *http.Request) {
	req.Header.Set("Referer", "https://wenshushu.cn/")
	req.Header.Set("User-Agent", "Chrome/80.0.3987.149 Wenshushu-Uploader")
	req.Header.Set("Origin", "https://wenshushu.cn/")
}

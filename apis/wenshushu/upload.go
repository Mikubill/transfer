package wenshushu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io/ioutil"
	"log"
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
	anonymous = "https://www.wenshushu.cn/ap/login/anonymous"
	addSend   = "https://www.wenshushu.cn/ap/task/addsend"
	getUpID   = "https://www.wenshushu.cn/ap/uploadv2/getupid"
	getUpURL  = "https://www.wenshushu.cn/ap/uploadv2/psurl"
	complete  = "https://www.wenshushu.cn/ap/uploadv2/complete"
	process   = "https://www.wenshushu.cn/ap/ufile/getprocess"
	finish    = "https://www.wenshushu.cn/ap/task/copysend"
)

func (b wssTransfer) Upload(files []string) {
	if b.Config.singleMode {
		b.initUpload(files)
	} else {
		for _, v := range files {
			b.initUpload([]string{v})
		}
	}
}

func (b wssTransfer) initUpload(files []string) {
	totalSize := int64(0)
	totalCount := 0
	for _, v := range files {
		if utils.IsExist(v) {
			err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				totalSize += info.Size()
				totalCount++
				return nil
			})
			if err != nil {
				fmt.Printf("filepath.walk returns error: %v, onfile: %s\n", err, v)
			}
		} else {
			fmt.Printf("%s not found\n", v)
		}
	}
	config, err := b.getSendConfig(totalSize, totalCount)
	if err != nil {
		fmt.Printf("getSendConfig(single mode) returns error: %v\n", err)
	}
	for _, v := range files {
		if utils.IsExist(v) {
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
		} else {
			fmt.Printf("%s not found\n", v)
		}
	}
	err = b.completeUpload(config)
	if err != nil {
		fmt.Printf("complete upload(single mode) returns error: %v\n", err)
	}
}

func (b wssTransfer) upload(v string, baseConf *sendConfigBlock) error {
	fmt.Printf("Local: %s\n", v)
	if b.Config.debugMode {
		log.Println("retrieving file info...")
	}
	info, err := utils.GetFileInfo(v)
	if err != nil {
		return fmt.Errorf("getFileInfo returns error: %v", err)
	}

	if info.Size()/int64(b.Config.blockSize) > 10000 {
		b.Config.blockSize = int(info.Size() / 10000)
		fmt.Printf("blocksize too small, set to %d\n", b.Config.blockSize)
	}

	bar := pb.Full.Start64(info.Size())
	bar.Set(pb.Bytes, true)
	file, err := os.Open(v)
	if err != nil {
		return fmt.Errorf("openFile returns error: %v", err)
	}

	wg := new(sync.WaitGroup)
	ch := make(chan *uploadPart)
	for i := 0; i < b.Config.parallel; i++ {
		go b.uploader(&ch, baseConf)
	}
	part := int64(0)
	for {
		part++
		buf := make([]byte, b.Config.blockSize)
		nr, err := file.Read(buf)
		if nr <= 0 || err != nil {
			break
		}
		if nr > 0 {
			wg.Add(1)
			ch <- &uploadPart{
				content: buf[:nr],
				count:   part,
				name:    v,
				wg:      wg,
				bar:     bar,
			}
		}
	}

	wg.Wait()
	close(ch)
	_ = file.Close()
	bar.Finish()
	// finish upload
	err = b.finishUpload(baseConf, info)
	if err != nil {
		return fmt.Errorf("finishUpload returns error: %v", err)
	}
	return nil
}

func (b wssTransfer) uploader(ch *chan *uploadPart, config *sendConfigBlock) {
	for item := range *ch {
		d, _ := json.Marshal(map[string]interface{}{
			"ispart": true,
			"fname":  item.name,
			"partnu": item.count,
			"fsize":  b.Config.blockSize,
			"upId":   config.UploadID,
		})
		uploadTicket, err := newRequest(getUpURL, string(d), requestConfig{
			debug:    b.Config.debugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(config.Token),
		})
		if err != nil {
			if b.Config.debugMode {
				log.Printf("get upload url request returns error: %v", err)
			}
			*ch <- item
			continue
		}

		client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
		data := new(bytes.Buffer)
		data.Write(item.content)
		if b.Config.debugMode {
			log.Printf("part %d start uploading", item.count)
			log.Printf("part %d posting %s", item.count, uploadTicket.Data.URL)
		}
		req, err := http.NewRequest("PUT", uploadTicket.Data.URL, data)
		if err != nil {
			if b.Config.debugMode {
				log.Printf("build request returns error: %v", err)
			}
			*ch <- item
			continue
		}
		req.Header.Set("content-type", "application/octet-stream")
		resp, err := client.Do(req)
		if err != nil {
			if b.Config.debugMode {
				log.Printf("failed uploading part %d error: %v (retrying)", item.count, err)
			}
			*ch <- item
			continue
		}
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			if b.Config.debugMode {
				log.Printf("failed uploading part %d error: %v (retrying)", item.count, err)
			}
			*ch <- item
			continue
		}

		_ = resp.Body.Close()

		if b.Config.debugMode {
			log.Printf("part %d finished.", item.count)
		}
		item.bar.Add(len(item.content))
		item.wg.Done()
	}

}

func (b wssTransfer) finishUpload(config *sendConfigBlock, info os.FileInfo) error {
	if b.Config.debugMode {
		log.Println("finish upload...")
		log.Println("step1 -> complete")
	}
	d, _ := json.Marshal(map[string]interface{}{
		"ispart": true,
		"fname":  info.Name(),
		"location": map[string]string{
			"boxid": config.Bid,
			"preid": config.UFileID,
		},
		"upId": config.UploadID,
	})

	body, err := newRequest(complete, string(d), requestConfig{
		debug:    b.Config.debugMode,
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

func (b wssTransfer) completeUpload(config *sendConfigBlock) error {
	if b.Config.debugMode {
		log.Println("complete upload...")
		log.Println("step1 -> process")
	}
	d, _ := json.Marshal(map[string]string{"processId": config.UploadID})
	for {
		body, err := newRequest(process, string(d), requestConfig{
			debug:    b.Config.debugMode,
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

	if b.Config.debugMode {
		log.Println("step2 -> finish(copySend)")
	}
	d, _ = json.Marshal(map[string]string{
		"bid":     config.Bid,
		"ufileid": config.UFileID,
		"tid":     config.Tid,
	})
	body, err := newRequest(finish, string(d), requestConfig{
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.Token),
	})
	if err != nil {
		return err
	}
	if body.Message != "success" {
		return fmt.Errorf("status != success")
	}

	fmt.Printf("Manage URL: %s\n", body.Data.ManageURL)
	fmt.Printf("Public URL: %s\n", body.Data.PublicURL)

	return nil
}

func (b wssTransfer) getTicket() (string, error) {
	if b.Config.token != "" {
		return b.Config.token, nil
	}
	if b.Config.debugMode {
		log.Println("getToken...")
	}
	config, err := newRequest(anonymous, "{\"dev_info\":\"{}\"}", requestConfig{
		debug:    b.Config.debugMode,
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

	if b.Config.debugMode {
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
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return nil, err
	}

	if b.Config.debugMode {
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
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return nil, err
	}
	config.Data.UploadID = upData.Data.UploadID
	config.Data.Token = ticket
	if b.Config.debugMode {
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
			return nil, fmt.Errorf("post %s returns error: ", err)
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
			return nil, fmt.Errorf("post %s returns error: ", err)
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
			return nil, fmt.Errorf("post %s returns error: ", err)
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
			return nil, fmt.Errorf("post %s returns error: ", err)
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

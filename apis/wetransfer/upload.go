package wetransfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"transfer/utils"
)

const (
	indexPage   = "https://wetransfer.com/"
	prepareSend = "https://wetransfer.com/api/v4/transfers/link"
	getSendURL  = "https://wetransfer.com/api/v4/transfers/%s/files"
	getUpURL    = "https://wetransfer.com/api/v4/transfers/%s/files/%s/part-put-url"
	finishPart  = "https://wetransfer.com/api/v4/transfers/%s/files/%s/finalize-mpp"
	finishAll   = "https://wetransfer.com/api/v4/transfers/%s/finalize"
	chunkSize   = 15728640
)

var tokenRegex = regexp.MustCompile("csrf-token\" content=\"([a-zA-Z0-9+=/]{88})\"")

func (b weTransfer) Upload(files []string) {
	if b.Config.singleMode {
		b.initUpload(files)
	} else {
		for _, v := range files {
			b.initUpload([]string{v})
		}
	}
}

func (b weTransfer) initUpload(files []string) {
	var fileItem []fileInfo
	for _, v := range files {
		if utils.IsExist(v) {
			err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				fileItem = append(fileItem, fileInfo{
					Name: info.Name(),
					Size: info.Size(),
					Type: "file",
				})
				return nil
			})
			if err != nil {
				fmt.Printf("filepath.walk returns error: %v, onfile: %s\n", err, v)
			}
		} else {
			fmt.Printf("%s not found\n", v)
		}
	}
	config, err := b.getSendConfig(fileItem)
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

func (b weTransfer) upload(v string, baseConf *configBlock) error {
	fmt.Printf("Local: %s\n", v)
	if b.Config.debugMode {
		log.Println("retrieving file info...")
	}
	info, err := utils.GetFileInfo(v)
	if err != nil {
		return fmt.Errorf("getFileInfo returns error: %v", err)
	}

	if b.Config.debugMode {
		log.Println("send file init...")
	}
	d, _ := json.Marshal(map[string]interface{}{
		"name": info.Name(),
		"size": info.Size(),
	})
	config, err := newRequest(fmt.Sprintf(getSendURL, baseConf.ID), string(d), requestConfig{
		action:   "POST",
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(baseConf.ticket),
	})
	if err != nil {
		return err
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
		buf := make([]byte, chunkSize)
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
				fileID:  config.ID,
			}
		}
	}

	wg.Wait()
	close(ch)
	_ = file.Close()
	bar.Finish()
	// finish upload
	err = b.finishUpload(baseConf, info, config.ID)
	if err != nil {
		return fmt.Errorf("finishUpload returns error: %v", err)
	}
	return nil
}

func (b weTransfer) uploader(ch *chan *uploadPart, config *configBlock) {
	for item := range *ch {
		d, _ := json.Marshal(map[string]interface{}{
			"chunk_number": item.count,
			"chunk_size":   len(item.content),
			"chunk_crc":    0,
		})
		uploadTicket, err := newRequest(fmt.Sprintf(getUpURL, config.ID, item.fileID), string(d), requestConfig{
			action:   "POST",
			debug:    b.Config.debugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(config.ticket),
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
			log.Printf("part %d posting %s", item.count, uploadTicket.URL)
		}
		req, err := http.NewRequest("PUT", uploadTicket.URL, data)
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

func (b weTransfer) finishUpload(config *configBlock, info os.FileInfo, id string) error {
	if b.Config.debugMode {
		log.Println("finish upload...")
		log.Println("step1 -> complete")
	}
	chunkCount := int(math.Round(float64(info.Size() / chunkSize)))
	d, _ := json.Marshal(map[string]interface{}{
		"chunk_count": chunkCount,
	})
	link := fmt.Sprintf(finishPart, config.ID, id)
	_, err := newRequest(link, string(d), requestConfig{
		action:   "PUT",
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.ticket),
	})
	if err != nil {
		return err
	}
	return nil
}

func (b weTransfer) completeUpload(config *configBlock) error {
	if b.Config.debugMode {
		log.Println("complete upload...")
		log.Println("step1 -> process")
	}
	link := fmt.Sprintf(finishAll, config.ID)
	body, err := newRequest(link, "", requestConfig{
		action:   "PUT",
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.ticket),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Public URL: %s\n", body.Public)
	return nil
}

func (b *weTransfer) getTicket() (requestTicket, error) {
	client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
	if b.Config.debugMode {
		log.Println("parse cookies, getToken...")
	}
	resp, err := client.Get(indexPage)
	if err != nil {
		return requestTicket{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return requestTicket{}, err
	}
	_ = resp.Body.Close()
	tk := tokenRegex.FindSubmatch(body)
	if b.Config.debugMode {
		log.Println("returns: ", string(tk[0]), string(tk[1]))
	}
	if len(tk) == 0 {
		return requestTicket{}, fmt.Errorf("no csrf-token found")
	}
	ticket := requestTicket{
		token:   string(tk[1]),
		cookies: resp.Header.Get("Set-Cookie"),
	}
	ck := resp.Header.Values("Set-Cookie")
	for _, v := range ck {
		s := strings.Split(v, ";")
		ticket.cookies += s[0] + ";"
	}
	return ticket, nil
}

func (b weTransfer) getSendConfig(info []fileInfo) (*configBlock, error) {
	fmt.Printf("fetching upload tickets..")
	end := utils.DotTicker()

	ticket, err := b.getTicket()
	if err != nil {
		return nil, err
	}

	if b.Config.debugMode {
		log.Println("step 1/2 email")
		log.Printf("ticket: %+v", ticket)
	}
	data, _ := json.Marshal(map[string]interface{}{
		"message":        "",
		"ui_language":    "en",
		"domain_user_id": utils.GenRandUUID(),
		"files":          info,
	})
	config, err := newRequest(prepareSend, string(data), requestConfig{
		action:   "POST",
		debug:    b.Config.debugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return nil, err
	}
	config.ticket.token = ticket.token
	if b.Config.debugMode {
		log.Println("init finished.")
		log.Printf("config: %+v", config)
	}
	*end <- struct{}{}
	fmt.Printf("ok\n")
	return config, nil
}

func newRequest(link string, postBody string, config requestConfig) (*configBlock, error) {
	if config.debug {
		log.Printf("endpoint: %s", link)
		log.Printf("postBody: %s", postBody)

	}

	client := http.Client{Timeout: config.timeout}
	req, err := http.NewRequest(config.action, link, strings.NewReader(postBody))
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

	respDat := new(configBlock)
	err = json.Unmarshal(body, respDat)
	if err != nil {
		if config.retry > 3 {
			return nil, fmt.Errorf("post %s returns error: ", err)
		}
		config.retry++
		return newRequest(link, postBody, config)
	}
	if config.debug {
		log.Printf("%+v", respDat)
	}
	ck := resp.Header.Values("Set-Cookie")
	for _, v := range ck {
		s := strings.Split(v, ";")
		respDat.ticket.cookies += s[0] + ";"
	}
	return respDat, nil
}

func addToken(token requestTicket) func(req *http.Request) {
	return func(req *http.Request) {
		addHeaders(req)
		req.Header.Set("x-csrf-token", token.token)
		req.Header.Set("cookie", token.cookies)
	}
}

func addHeaders(req *http.Request) {
	req.Header.Set("Referer", "https://wetransfer.com/")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("User-Agent", "Chrome/80.0.3987.149 Wenshushu-Uploader")
	req.Header.Set("Origin", "https://wetransfer.com/")
}

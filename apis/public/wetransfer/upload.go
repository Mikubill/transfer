package wetransfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/cheggaaa/pb/v3"
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

var tokenRegex = regexp.MustCompile("csrf-token\"\\scontent=\"([^\"]+)\"\\s/>?")

func (b *weTransfer) InitUpload(files []string, sizes []int64) error {
	if b.Config.singleMode {
		var fileItem []fileInfo
		for n, v := range files {
			fileItem = append(fileItem, fileInfo{
				Name: filepath.Base(v),
				Size: sizes[n],
				Type: "file",
			})
		}
		b.initUpload(fileItem)
	}
	return nil
}

func (b *weTransfer) StartProgress(reader io.Reader, size int64) io.Reader {
	bar := pb.Full.Start64(size)
	bar.Set(pb.Bytes, true)
	b.Bar = bar
	return reader
}

func (b *weTransfer) initUpload(fileItem []fileInfo) {

	err := b.getSendConfig(fileItem)
	if err != nil {
		fmt.Printf("getSendConfig error: %v\n", err)
	}
}

func (b *weTransfer) PreUpload(name string, size int64) error {
	if !b.Config.singleMode {
		b.initUpload([]fileInfo{
			{
				Name: name,
				Size: size,
				Type: "file",
			},
		})
	}
	return nil
}

func (b weTransfer) DoUpload(name string, size int64, file io.Reader) error {

	if apis.DebugMode {
		log.Println("send file init...")
	}
	d, _ := json.Marshal(map[string]any{
		"name": name,
		"size": size,
	})
	config, err := newRequest(fmt.Sprintf(getSendURL, b.baseConf.ID), string(d), requestConfig{
		action:   "POST",
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(b.baseConf.ticket),
	})
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	ch := make(chan *uploadPart)
	for i := 0; i < b.Config.Parallel; i++ {
		go b.uploader(&ch, b.baseConf)
	}
	part := int64(0)
	for {
		part++
		buf := make([]byte, chunkSize)
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
				fileID:  config.ID,
			}
		}
	}

	wg.Wait()
	close(ch)
	// finish upload
	err = b.finishUpload(b.baseConf, size, config.ID)
	if err != nil {
		return fmt.Errorf("finishUpload returns error: %v", err)
	}

	return nil
}

func (b weTransfer) PostUpload(string, int64) (string, error) {
	if !b.Config.singleMode {
		return b.completeUpload(b.baseConf)
	}
	return "", nil
}

func (b weTransfer) FinishUpload([]string) (string, error) {
	if b.Config.singleMode {
		return b.completeUpload(b.baseConf)
	}
	return "", nil
}

func (b weTransfer) uploader(ch *chan *uploadPart, config *configBlock) {
	for item := range *ch {
	Start:
		d, _ := json.Marshal(map[string]any{
			"chunk_number": item.count,
			"chunk_size":   len(item.content),
			"chunk_crc":    0,
		})
		uploadTicket, err := newRequest(fmt.Sprintf(getUpURL, config.ID, item.fileID), string(d), requestConfig{
			action:   "POST",
			debug:    apis.DebugMode,
			retry:    0,
			timeout:  time.Duration(b.Config.interval) * time.Second,
			modifier: addToken(config.ticket),
		})
		if err != nil {
			if apis.DebugMode {
				log.Printf("get upload url request returns error: %v", err)
			}
			goto Start
		}

		client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
		data := new(bytes.Buffer)
		data.Write(item.content)
		if apis.DebugMode {
			log.Printf("part %d start uploading", item.count)
			log.Printf("part %d posting %s", item.count, uploadTicket.URL)
		}
		reader, writer := io.Pipe()
		go func() { _, _ = io.Copy(writer, data) }()
		var req *http.Request
		if b.Bar != nil {
			req, err = http.NewRequest("PUT", uploadTicket.URL, b.Bar.NewProxyReader(reader))
		} else {
			req, err = http.NewRequest("PUT", uploadTicket.URL, reader)
		}
		if err != nil {
			if apis.DebugMode {
				log.Printf("build request returns error: %v", err)
			}
			goto Start
		}
		req.ContentLength = int64(len(item.content))
		req.Header.Set("content-type", "application/octet-stream")
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if apis.DebugMode {
				log.Printf("failed uploading part %d error: %v (retrying)", item.count, err)
			}
			goto Start
		}
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			if apis.DebugMode {
				log.Printf("failed uploading part %d error: %v (retrying)", item.count, err)
			}
			goto Start
		}

		_ = resp.Body.Close()

		if apis.DebugMode {
			log.Printf("part %d finished.", item.count)
		}
		item.wg.Done()
	}

}

func (b weTransfer) finishUpload(config *configBlock, size int64, id string) error {
	if apis.DebugMode {
		log.Println("finish upload...")
		log.Println("step1 -> complete")
	}
	chunkCount := int(math.Ceil(float64(size) / float64(chunkSize)))
	d, _ := json.Marshal(map[string]any{
		"chunk_count": chunkCount,
	})
	link := fmt.Sprintf(finishPart, config.ID, id)
	_, err := newRequest(link, string(d), requestConfig{
		action:   "PUT",
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.ticket),
	})
	if err != nil {
		return err
	}
	return nil
}

func (b weTransfer) completeUpload(config *configBlock) (string, error) {
	if apis.DebugMode {
		log.Println("complete upload...")
		log.Println("step1 -> process")
	}
	link := fmt.Sprintf(finishAll, config.ID)
	body, err := newRequest(link, "", requestConfig{
		action:   "PUT",
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(config.ticket),
	})
	if err != nil {
		return "", err
	}
	fmt.Printf("Download Link: %s\n", body.Public)
	return body.Public, nil
}

func (b *weTransfer) getTicket() (requestTicket, error) {
	client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
	if apis.DebugMode {
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
	if apis.DebugMode {
		// parsedBody := strings.TrimSpace(string(body))
		// parsedBody = strings.Trim(parsedBody, "\n")
		// log.Println("returns: ", parsedBody)
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

func (b *weTransfer) getSendConfig(info []fileInfo) error {
	fmt.Printf("fetching upload tickets..")
	end := utils.DotTicker()

	ticket, err := b.getTicket()
	if err != nil {
		return err
	}

	if apis.DebugMode {
		log.Println("step 1/2 email")
		log.Printf("ticket: %+v", ticket)
	}
	data, _ := json.Marshal(map[string]any{
		"message":        "",
		"ui_language":    "en",
		"domain_user_id": utils.GenRandUUID(),
		"files":          info,
	})
	config, err := newRequest(prepareSend, string(data), requestConfig{
		action:   "POST",
		debug:    apis.DebugMode,
		retry:    0,
		timeout:  time.Duration(b.Config.interval) * time.Second,
		modifier: addToken(ticket),
	})
	if err != nil {
		return err
	}
	config.ticket.token = ticket.token
	if apis.DebugMode {
		log.Println("init finished.")
		log.Printf("config: %+v", config)
	}
	*end <- struct{}{}
	fmt.Printf("ok\n")
	b.baseConf = config
	return nil
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
			return nil, fmt.Errorf("post %s returns error: %s", link, err)
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

	respDat := new(configBlock)
	err = json.Unmarshal(body, respDat)
	if err != nil {
		if config.retry > 3 {
			return nil, fmt.Errorf("post %s returns error: %s", link, err)
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
		req.Header.Set("x-requested-with", "XMLHttpRequest")
		req.Header.Set("x-csrf-token", token.token)
		req.Header.Set("cookie", token.cookies)
	}
}

func addHeaders(req *http.Request) {
	req.Header.Set("Referer", "https://wetransfer.com/")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("User-Agent", "Chrome/80.0.3987.149 Wetransfer-Uploader")
	req.Header.Set("Origin", "https://wetransfer.com/")
}

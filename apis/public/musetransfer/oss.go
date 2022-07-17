package musetransfer

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Mikubill/transfer/apis"
)

func (b muse) uploader(ch *chan *uploadPart) {
	var err error
	for item := range *ch {
		// calc md5
		md5er := md5.New()
		if _, err = md5er.Write(item.content); err != nil {
			log.Printf("calc md5 error: %s", err)
			continue
		}
		md5sum := md5er.Sum(nil)
		encodedMD5 := base64.StdEncoding.EncodeToString(md5sum)

	Start:
		client := http.Client{Timeout: time.Duration(b.Config.interval) * time.Second}
		data := new(bytes.Buffer)
		data.Write(item.content)
		if apis.DebugMode {
			log.Printf("part %d start uploading", item.count)
			log.Printf("part %d posting %s", item.count, item.dest)
		}
		reader, writer := io.Pipe()
		go func() { _, _ = io.Copy(writer, data) }()

		var req *http.Request
		if b.Bar != nil {
			req, err = http.NewRequest("PUT", item.dest, b.Bar.NewProxyReader(reader))
		} else {
			req, err = http.NewRequest("PUT", item.dest, reader)
		}
		if err != nil {
			if apis.DebugMode {
				log.Printf("build request returns error: %v", err)
			}
			goto Start
		}
		req.ContentLength = int64(len(item.content))
		req.Header.Set("content-type", "application/octet-stream")
		b.ossSigner(req, item.auth, "application/octet-stream", encodedMD5)
		resp, err := client.Do(req)
		if err != nil {
			if apis.DebugMode {
				log.Printf("failed uploading part %d error: %v (retrying)", item.count, err)
				// log.Printf("last Resp: %v", res p)
			}
			goto Start
		}
		_ = resp.Body.Close()
		if apis.DebugMode {
			// print resp
			log.Printf("part %d finished.", item.count)
		}
		b.EtagMap.Store(item.count, resp.Header.Get("etag"))
		item.wg.Done()
	}

}

func (b muse) postAPI(ep string, data []byte) ([]byte, error) {
	k := bytes.NewBuffer(data)
	if apis.DebugMode {
		log.Println("\nrequesting", ep)
	}
	req, err := http.NewRequest("POST", ep, k)
	if err != nil {
		return nil, err
	}
	addToken(req, b.Config.devicetoken, ep, data)

	http.DefaultClient.Timeout = time.Duration(b.Config.interval) * time.Second
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if apis.DebugMode {
		log.Println("\nresponse:", string(body))
	}
	return body, nil
}

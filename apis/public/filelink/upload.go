package filelink

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const upload = "https://filelink.txthinking.com/"

func (b *fileLink) DoUpload(name string, size int64, file io.Reader) error {

	body, err := b.newUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      b.Config.DebugMode,
	})
	if err != nil {
		return fmt.Errorf("upload error: %s", err)
	}
	//fmt.Println("Download Link: " + string(body))
	b.result = "http://i.filelink.io/" + string(body)
	return nil
}

func (b fileLink) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: %s\n", b.result)
	return b.result, nil
}

func (b fileLink) newUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	req, err := http.NewRequest("POST", upload, config.fileReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/octet-stream")
	req.Header.Set("referer", "https://filelink.io/")
	req.Header.Set("origin", "https://filelink.io/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh;)")
	req.Header.Set("x-file-name", config.fileName)
	if config.debug {
		log.Printf("header: %v", req.Header)
	}
	resp, err := client.Do(req)
	if err != nil {
		if config.debug {
			log.Printf("do requests returns error: %v", err)
		}
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", err)
		}
		return nil, err
	}
	_ = resp.Body.Close()
	if config.debug {
		log.Printf("returns: %v", string(body))
	}

	return body, nil
}

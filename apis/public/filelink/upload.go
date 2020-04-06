package filelink

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const upload = "https://filelink.io/up"

func (b fileLink) DoUpload(name string, size int64, file io.Reader) error {

	body, err := b.newUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      b.Config.DebugMode,
	})
	if err != nil {
		return fmt.Errorf("upload error: %s", err)
	}
	fmt.Println("Download Link: " + string(body))

	return nil
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
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
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

package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

// Command Types
const (
	signedURLPrefix = "https://www.notion.so/signed"
	notionHost      = "https://www.notion.so"
)

func PrintStruct(emp any) {
	empJSON, err := json.MarshalIndent(emp, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("MarshalIndent funnction output\n %s\n", string(empJSON))
}

func (b *notion) DoUpload(name string, size int64, file io.Reader) error {

	if b.pageID == "" || b.token == "" {
		return fmt.Errorf("invalid pageid or token")
	}
	client := NewWebClient(b.token)
	root, err := client.GetPage(b.pageID)
	if err != nil {
		log.Fatalf("GetPage() failed with %s\n", err)
	}
	if apis.DebugMode {
		PrintStruct(root)
	}

	fileID, fileURL, err := client.UploadFile(file, name, size)
	if err != nil {
		log.Fatalf("UploadFile() failed with %s\n", err)
	}
	if apis.DebugMode {
		log.Printf("id: %s, url: %s", fileID, fileURL)
	}
	fmt.Printf("syncing blocks..")
	end := utils.DotTicker()
	newBlockID, err := client.insertFile(name, fileID, fileURL, root, size, time.Now())
	if err != nil {
		log.Fatalf("insertFile() failed with %s\n", err)
	}

	*end <- struct{}{}
	fmt.Printf("%s\n", newBlockID)
	b.resp = fmt.Sprintf("%s/%s?table=block&id=%s&name=%s&userId=%s&cache=v2", signedURLPrefix, url.QueryEscape(fileURL), newBlockID, name, root.Owneruserid)
	return nil
}

func (b notion) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: %s\n", b.resp)
	return b.resp, nil
}
